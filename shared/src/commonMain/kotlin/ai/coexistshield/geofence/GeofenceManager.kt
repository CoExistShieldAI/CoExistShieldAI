package ai.coexistshield.geofence

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.distanceToKm
import ai.coexistshield.model.distanceToMeters
import ai.coexistshield.network.GeofenceEventDto
import ai.coexistshield.repository.GeofenceRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * Abstraction layer for native platform geofence registration (iOS CoreLocation / Android GeofencingClient).
 */
interface NativeGeofenceRegistrar {
    suspend fun registerGeofences(zones: List<GeofenceZone>)
    suspend fun unregisterGeofences(zoneIds: List<String>)
    fun getRegisteredZoneIds(): Set<String>
}

/**
 * In-memory test/fallback implementation of NativeGeofenceRegistrar.
 */
class InMemoryNativeGeofenceRegistrar : NativeGeofenceRegistrar {
    private val activeZoneMap = mutableMapOf<String, GeofenceZone>()

    override suspend fun registerGeofences(zones: List<GeofenceZone>) {
        for (z in zones) {
            activeZoneMap[z.id] = z
        }
    }

    override suspend fun unregisterGeofences(zoneIds: List<String>) {
        for (id in zoneIds) {
            activeZoneMap.remove(id)
        }
    }

    override fun getRegisteredZoneIds(): Set<String> = activeZoneMap.keys.toSet()

    fun getRegisteredZones(): List<GeofenceZone> = activeZoneMap.values.toList()
}

enum class GeofenceSyncState {
    IDLE,
    SYNCING,
    SYNCED,
    CONNECTED,
    ERROR
}

/**
 * Manages spatial filtering, quota compliance (max 20 geofences for iOS, max 100 for Android),
 * movement hysteresis, and immediate synchronization with backend geofence mutations.
 */
class GeofenceManager(
    private val repository: GeofenceRepository,
    private val registrar: NativeGeofenceRegistrar = InMemoryNativeGeofenceRegistrar(),
    val maxGeofences: Int = 20, // Default to iOS strict limit of 20
    val movementThresholdMeters: Double = 500.0, // Hysteresis threshold to prevent battery drain & chattering
    val searchRadiusKm: Double = 50.0 // Regional search window
) {
    private val _registeredGeofences = MutableStateFlow<List<GeofenceZone>>(emptyList())
    val registeredGeofences: StateFlow<List<GeofenceZone>> = _registeredGeofences.asStateFlow()

    private val _allCachedGeofences = MutableStateFlow<List<GeofenceZone>>(emptyList())
    val allCachedGeofences: StateFlow<List<GeofenceZone>> = _allCachedGeofences.asStateFlow()

    private val _syncState = MutableStateFlow(GeofenceSyncState.IDLE)
    val syncState: StateFlow<GeofenceSyncState> = _syncState.asStateFlow()

    private val _lastSyncedAt = MutableStateFlow<String?>(null)
    val lastSyncedAt: StateFlow<String?> = _lastSyncedAt.asStateFlow()

    private val _userLocation = MutableStateFlow<GeoLocation?>(null)
    val userLocation: StateFlow<GeoLocation?> = _userLocation.asStateFlow()

    private var lastEvaluatedLocation: GeoLocation? = null
    private var syncJob: Job? = null

    /**
     * Start background real-time synchronization with backend Server-Sent Events (SSE).
     * Any change in the backend (AI recalculation, manual zone added/deleted) triggers immediate local reconciliation.
     */
    fun startRealtimeSync(scope: CoroutineScope): Job {
        syncJob?.cancel()
        val job = scope.launch {
            _syncState.value = GeofenceSyncState.CONNECTED
            repository.observeGeofenceEvents().collect { event ->
                handleBackendGeofenceEvent(event)
            }
        }
        syncJob = job
        return job
    }

    fun stopRealtimeSync() {
        syncJob?.cancel()
        syncJob = null
        _syncState.value = GeofenceSyncState.IDLE
    }

    /**
     * Handles an incoming real-time geofence event from backend.
     */
    suspend fun handleBackendGeofenceEvent(event: GeofenceEventDto) {
        when (event.type) {
            "CONNECTED" -> {
                _syncState.value = GeofenceSyncState.CONNECTED
            }
            "RECALCULATED", "CREATED", "UPDATED", "DELETED" -> {
                // If user location is known, immediately refresh proximity geofences
                _userLocation.value?.let { loc ->
                    refreshGeofences(loc, forceUpdate = true)
                }
            }
        }
    }

    /**
     * Called whenever device location updates.
     * Evaluates whether device moved past the hysteresis threshold before updating native geofences.
     */
    suspend fun onLocationChanged(newLocation: GeoLocation, forceRefresh: Boolean = false) {
        _userLocation.value = newLocation

        val lastLoc = lastEvaluatedLocation
        if (!forceRefresh && lastLoc != null) {
            val dist = newLocation.distanceToMeters(lastLoc)
            if (dist < movementThresholdMeters) {
                // Movement within hysteresis threshold; skip costly OS geofence re-registration
                return
            }
        }

        refreshGeofences(newLocation, forceUpdate = forceRefresh)
    }

    /**
     * Fetches candidate geofences around device and applies threat-proximity ranking
     * to keep registered zones within OS limits (e.g. max 20).
     */
    suspend fun refreshGeofences(location: GeoLocation, forceUpdate: Boolean = false) {
        _syncState.value = GeofenceSyncState.SYNCING

        val result = repository.getActiveGeofences(
            latitude = location.latitude,
            longitude = location.longitude,
            radiusKm = searchRadiusKm
        )

        result.fold(
            onSuccess = { fetchedZones ->
                _allCachedGeofences.value = fetchedZones
                lastEvaluatedLocation = location

                // Smart filtering: Rank zones based on Threat Tier & Proximity
                val prioritizedZones = filterAndRankGeofences(fetchedZones, location, maxGeofences)
                reconcileNativeRegistrations(prioritizedZones)

                _registeredGeofences.value = prioritizedZones
                _syncState.value = GeofenceSyncState.SYNCED
                _lastSyncedAt.value = "Synced (${prioritizedZones.size}/$maxGeofences active)"
            },
            onFailure = {
                _syncState.value = GeofenceSyncState.ERROR
            }
        )
    }

    /**
     * Ranks geofences using a threat-weighted distance decay function:
     * Score = TierWeight / (1.0 + DistanceKm)
     * High danger tier (TIER_3_DANGER) remains prioritized even at moderate distance,
     * while distant or low-priority zones are safely pruned to satisfy OS constraints.
     */
    fun filterAndRankGeofences(
        zones: List<GeofenceZone>,
        userLocation: GeoLocation,
        limit: Int
    ): List<GeofenceZone> {
        return zones
            .map { zone ->
                val distKm = userLocation.distanceToKm(zone.centroid)
                val weight = when (zone.tier) {
                    AlertTier.TIER_3_DANGER -> 100.0
                    AlertTier.TIER_2_WARNING -> 50.0
                    AlertTier.TIER_1_NOTICE -> 20.0
                    AlertTier.SAFE -> 5.0
                }
                val score = weight / (1.0 + distKm)
                Pair(zone, score)
            }
            .sortedByDescending { it.second }
            .take(limit)
            .map { it.first }
    }

    /**
     * Differential synchronization with the underlying native OS registrar.
     * Only unregisters zones no longer in the priority set, and registers newly promoted ones.
     */
    private suspend fun reconcileNativeRegistrations(targetZones: List<GeofenceZone>) {
        val currentRegisteredIds = registrar.getRegisteredZoneIds()
        val targetIds = targetZones.map { it.id }.toSet()

        val toRemove = currentRegisteredIds - targetIds
        val toAdd = targetZones.filter { it.id !in currentRegisteredIds }

        if (toRemove.isNotEmpty()) {
            registrar.unregisterGeofences(toRemove.toList())
        }
        if (toAdd.isNotEmpty()) {
            registrar.registerGeofences(toAdd)
        }
    }

    /**
     * Evaluates the graduated alert tier for a given location using both direct containment
     * and spatial concentric buffer halos. Ensures that exiting or approaching a Tier 3 Critical zone
     * naturally scales through Tier 2 Warning and Tier 1 Advisory stages rather than abruptly jumping to Safe.
     *
     * Returns Triple(AlertTier, minDistanceToBoundaryMeters, matchedZoneId)
     */
    fun evaluateGraduatedTier(
        location: GeoLocation,
        candidateZones: List<GeofenceZone> = _allCachedGeofences.value
    ): Triple<AlertTier, Double, String?> {
        val tierRank = mapOf(
            AlertTier.SAFE to 0,
            AlertTier.TIER_1_NOTICE to 1,
            AlertTier.TIER_2_WARNING to 2,
            AlertTier.TIER_3_DANGER to 3
        )

        var highestTier = AlertTier.SAFE
        var minDistMeters = Double.MAX_VALUE
        var bestZoneId: String? = null

        for (zone in candidateZones) {
            val centerDist = location.distanceToMeters(zone.centroid)
            val effectiveBoundaryDist = kotlin.math.max(0.0, centerDist - zone.radiusMeters)

            if (effectiveBoundaryDist < minDistMeters) {
                minDistMeters = effectiveBoundaryDist
                bestZoneId = zone.id
            }

            if (centerDist <= zone.radiusMeters) {
                if ((tierRank[zone.tier] ?: 0) > (tierRank[highestTier] ?: 0)) {
                    highestTier = zone.tier
                }
            } else {
                // Graduated concentric buffer halo evaluation
                when (zone.tier) {
                    AlertTier.TIER_3_DANGER -> {
                        val t2Threshold = kotlin.math.max(1200.0, zone.radiusMeters * 0.75)
                        val t1Threshold = kotlin.math.max(2500.0, zone.radiusMeters * 1.60)
                        if (effectiveBoundaryDist <= t2Threshold) {
                            if ((tierRank[AlertTier.TIER_2_WARNING] ?: 0) > (tierRank[highestTier] ?: 0)) {
                                highestTier = AlertTier.TIER_2_WARNING
                            }
                        } else if (effectiveBoundaryDist <= t1Threshold) {
                            if ((tierRank[AlertTier.TIER_1_NOTICE] ?: 0) > (tierRank[highestTier] ?: 0)) {
                                highestTier = AlertTier.TIER_1_NOTICE
                            }
                        }
                    }
                    AlertTier.TIER_2_WARNING -> {
                        val t1Threshold = kotlin.math.max(1500.0, zone.radiusMeters * 0.75)
                        if (effectiveBoundaryDist <= t1Threshold) {
                            if ((tierRank[AlertTier.TIER_1_NOTICE] ?: 0) > (tierRank[highestTier] ?: 0)) {
                                highestTier = AlertTier.TIER_1_NOTICE
                            }
                        }
                    }
                    else -> Unit
                }
            }
        }

        if (minDistMeters == Double.MAX_VALUE) {
            minDistMeters = 0.0
        }

        return Triple(highestTier, minDistMeters, bestZoneId)
    }
}
