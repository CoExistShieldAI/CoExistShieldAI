package ai.coexistshield.ui.map

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.ClearanceStatus
import ai.coexistshield.model.DecayState
import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.RoadkillRecord
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.ui.theme.TierDanger
import ai.coexistshield.ui.theme.TierNotice
import ai.coexistshield.ui.theme.TierSafe
import ai.coexistshield.ui.theme.TierWarning
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.FloatingActionButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Fill
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.IntSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.CoroutineScope
import kotlin.math.pow
import kotlin.math.roundToInt
import kotlin.math.sin

object MapZoomHelper {
    const val MIN_ZOOM = 5.0f // Nationwide overview (~1000km)
    const val MAX_ZOOM = 15.0f // Street & parcel level (~100m)
    const val DEFAULT_ZOOM = 13.0f // City / district level (~1km)

    data class ZoomMetrics(
        val discreteZoom: Int,
        val scale: Float,
        val tileSizePx: Float
    )

    fun calculateMetrics(zoomLevel: Float): ZoomMetrics {
        val clampedZoom = zoomLevel.coerceIn(MIN_ZOOM, MAX_ZOOM)
        val discreteZoom = clampedZoom.toInt().coerceIn(5, 15)
        // Fractional part provides smooth continuous scaling: 2^(zoom - discreteZoom)
        val frac = clampedZoom - discreteZoom
        val scale = 2.0f.pow(frac)
        val tileSizePx = 256f * scale
        return ZoomMetrics(discreteZoom, scale, tileSizePx)
    }
}

enum class MapLayerType(val label: String) {
    GEOFENCES("Dynamic Geofences"),
    ROADKILL("Roadkill"),
    PERSIMMON_TREES("Kaki Trees (Japan)"),
    ELEPHANT_CROPS("Elephant Crops (Thai)"),
    LIVESTOCK_PREDATION("Tiger Corrals / Kills")
}

data class SelectedEntityDetails(
    val id: String,
    val title: String,
    val subtitle: String,
    val speciesOrType: String,
    val threatTier: AlertTier = AlertTier.TIER_1_NOTICE,
    val geminiBadgeText: String? = null,
    val canMarkResolved: Boolean = false,
    val isRoadkill: Boolean = false,
    val roadkillStatus: ClearanceStatus? = null,
    val isGeofence: Boolean = false,
    val radiusMeters: Double? = null,
    val latitude: Double? = null,
    val longitude: Double? = null
)

data class GisMapUiState(
    val userLocation: GeoLocation? = GeoLocation(39.702, 141.152),
    val activeGeofences: List<GeofenceZone> = emptyList(),
    val roadkills: List<RoadkillRecord> = emptyList(),
    val proximityRoadkillAlert: RoadkillRecord? = null,
    val activeLayers: Set<MapLayerType> = setOf(
        MapLayerType.GEOFENCES,
        MapLayerType.ROADKILL,
        MapLayerType.PERSIMMON_TREES,
        MapLayerType.ELEPHANT_CROPS,
        MapLayerType.LIVESTOCK_PREDATION
    ),
    val selectedEntity: SelectedEntityDetails? = null,
    val userRole: UserRole = UserRole.CITIZEN
)

enum class MapRegion(val label: String, val centerLat: Double, val centerLon: Double) {
    JAPAN_IWATE("🇯🇵 Iwate, Japan", 39.702, 141.152),
    THAILAND_KUI_BURI("🇹🇭 Kui Buri, Thailand", 12.052, 99.821),
    INDONESIA_SUMATRA("🇮🇩 Bukit Kesuma, Sumatra", 0.125, 101.450)
}

data class MapPinItem(
    val id: String,
    val lat: Double,
    val lon: Double,
    val iconEmoji: String,
    val title: String,
    val subtitle: String,
    val layerType: MapLayerType,
    val threatTier: AlertTier,
    val pinColor: Color,
    val geminiBadge: String,
    val isRoadkill: Boolean = false,
    val roadkillStatus: ClearanceStatus? = null,
    val isGeofence: Boolean = false,
    val radiusMeters: Double? = null
)

sealed interface MapDisplayItem {
    val screenX: Float
    val screenY: Float
}

data class SinglePinDisplay(
    val pin: MapPinItem,
    override val screenX: Float,
    override val screenY: Float
) : MapDisplayItem

data class ClusterDisplay(
    val pins: List<MapPinItem>,
    override val screenX: Float,
    override val screenY: Float,
    val highestTier: AlertTier,
    val badgeColor: Color,
    val centroidLat: Double,
    val centroidLon: Double
) : MapDisplayItem

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GisHazardMapScreen(
    state: GisMapUiState,
    initialRegion: MapRegion? = null,
    initialTargetLocation: GeoLocation? = null,
    initialSelectedEntityId: String? = null,
    onBackClick: () -> Unit = {},
    onToggleLayer: (MapLayerType) -> Unit = {},
    onSelectEntity: (SelectedEntityDetails?) -> Unit = {},
    onOfficialActionClick: (String) -> Unit = {},
    onUpdateRoadkillStatus: (String, ClearanceStatus) -> Unit = { _, _ -> }
) {
    var selectedRegion by remember { mutableStateOf(initialRegion ?: MapRegion.JAPAN_IWATE) }
    var panOffsetX by remember { mutableStateOf(0f) }
    var panOffsetY by remember { mutableStateOf(0f) }
    var zoomLevel by remember { mutableStateOf(MapZoomHelper.DEFAULT_ZOOM) }

    LaunchedEffect(initialTargetLocation, initialSelectedEntityId, initialRegion) {
        val targetReg = initialRegion ?: initialTargetLocation?.let { loc ->
            when {
                loc.latitude in 25.0..50.0 -> MapRegion.JAPAN_IWATE
                loc.latitude in 5.0..25.0 -> MapRegion.THAILAND_KUI_BURI
                else -> MapRegion.INDONESIA_SUMATRA
            }
        }
        if (targetReg != null) {
            selectedRegion = targetReg
        }

        if (initialTargetLocation != null) {
            val reg = targetReg ?: selectedRegion
            val targetZoom = 13.5f
            val metrics = MapZoomHelper.calculateMetrics(targetZoom)
            val centerTileX = TileCoordinateUtils.lonToTileX(reg.centerLon, metrics.discreteZoom)
            val centerTileY = TileCoordinateUtils.latToTileY(reg.centerLat, metrics.discreteZoom)
            val targetTileX = TileCoordinateUtils.lonToTileX(initialTargetLocation.longitude, metrics.discreteZoom)
            val targetTileY = TileCoordinateUtils.latToTileY(initialTargetLocation.latitude, metrics.discreteZoom)
            panOffsetX = -((targetTileX - centerTileX).toFloat() * metrics.tileSizePx)
            panOffsetY = -((targetTileY - centerTileY).toFloat() * metrics.tileSizePx)
            zoomLevel = targetZoom
        }

        if (initialSelectedEntityId != null) {
            val geofence = state.activeGeofences.firstOrNull { it.id == initialSelectedEntityId }
            if (geofence != null) {
                onSelectEntity(
                    SelectedEntityDetails(
                        id = geofence.id,
                        title = geofence.causalExplanation.substringBefore(":").trim().ifEmpty { "Active Geofence Shield" },
                        subtitle = geofence.causalExplanation,
                        speciesOrType = geofence.targetSpecies.firstOrNull()?.name ?: "UNKNOWN",
                        threatTier = geofence.tier,
                        geminiBadgeText = "Gemini Spatial: Active Hazard Zone • Radius ${geofence.radiusMeters?.toInt() ?: 1500}m",
                        canMarkResolved = false,
                        isGeofence = true,
                        radiusMeters = geofence.radiusMeters,
                        latitude = geofence.centroid?.latitude,
                        longitude = geofence.centroid?.longitude
                    )
                )
            } else {
                val rk = state.roadkills.firstOrNull { it.id == initialSelectedEntityId }
                if (rk != null) {
                    onSelectEntity(
                        SelectedEntityDetails(
                            id = rk.id,
                            title = "Decomposing ${rk.species.name.replace("_", " ")} Carcass",
                            subtitle = "Active Apex Predator Lure • Attractant Threat Score ${rk.attractantThreatScore}",
                            speciesOrType = rk.species.name,
                            threatTier = AlertTier.TIER_2_WARNING,
                            geminiBadgeText = "Gemini Forensics: ${rk.decayState.name} • Threat ${rk.attractantThreatScore}",
                            canMarkResolved = true,
                            isRoadkill = true,
                            roadkillStatus = rk.clearanceStatus,
                            latitude = rk.location.latitude,
                            longitude = rk.location.longitude
                        )
                    )
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("GIS Hazard Map", fontSize = 17.sp, fontWeight = FontWeight.Bold, color = Color.White)
                        Spacer(modifier = Modifier.width(6.dp))
                        Box(
                            modifier = Modifier
                                .background(Color(0xFF38BDF8).copy(alpha = 0.2f), RoundedCornerShape(4.dp))
                                .border(1.dp, Color(0xFF38BDF8), RoundedCornerShape(4.dp))
                                .padding(horizontal = 6.dp, vertical = 2.dp)
                        ) {
                            Text("LIVE TILES", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = Color(0xFF38BDF8))
                        }
                    }
                },
                navigationIcon = {
                    Box(modifier = Modifier.padding(start = 12.dp)) {
                        OutlinedButton(
                            onClick = onBackClick,
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text("HUD", fontSize = 12.sp)
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = Color(0xFF1E293B)
                )
            )
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
        ) {
            // 1. Live Web Mercator Raster Tile & Dynamic GeoJSON Canvas Engine
            GoogleMapsVectorCanvas(
                region = selectedRegion,
                state = state,
                panOffsetX = panOffsetX,
                panOffsetY = panOffsetY,
                zoomLevel = zoomLevel,
                onPanChange = { dx, dy ->
                    panOffsetX += dx
                    panOffsetY += dy
                },
                onSelectPin = { pin ->
                    onSelectEntity(
                        SelectedEntityDetails(
                            id = pin.id,
                            title = pin.title,
                            subtitle = pin.subtitle,
                            speciesOrType = pin.layerType.name,
                            threatTier = pin.threatTier,
                            geminiBadgeText = pin.geminiBadge,
                            canMarkResolved = true,
                            isRoadkill = pin.isRoadkill,
                            roadkillStatus = pin.roadkillStatus,
                            isGeofence = pin.isGeofence,
                            radiusMeters = pin.radiusMeters,
                            latitude = pin.lat,
                            longitude = pin.lon
                        )
                    )
                },
                onClusterClick = { cluster ->
                    // Center map on cluster centroid and zoom in to expand pins smoothly
                    val newZoom = (zoomLevel + 2.5f).coerceAtMost(MapZoomHelper.MAX_ZOOM)
                    val newMetrics = MapZoomHelper.calculateMetrics(newZoom)
                    val centerTileX = TileCoordinateUtils.lonToTileX(selectedRegion.centerLon, newMetrics.discreteZoom)
                    val centerTileY = TileCoordinateUtils.latToTileY(selectedRegion.centerLat, newMetrics.discreteZoom)
                    val targetTileX = TileCoordinateUtils.lonToTileX(cluster.centroidLon, newMetrics.discreteZoom)
                    val targetTileY = TileCoordinateUtils.latToTileY(cluster.centroidLat, newMetrics.discreteZoom)

                    panOffsetX = -((targetTileX - centerTileX).toFloat() * newMetrics.tileSizePx)
                    panOffsetY = -((targetTileY - centerTileY).toFloat() * newMetrics.tileSizePx)
                    zoomLevel = newZoom
                }
            )

            // 2. Top Region Selector & Layer Filter Bar Overlay
            Column(
                modifier = Modifier
                    .align(Alignment.TopCenter)
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                // Region Quick-Jump Tabs
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    MapRegion.values().forEach { region ->
                        val isSelected = selectedRegion == region
                        Box(
                            modifier = Modifier
                                .clip(RoundedCornerShape(16.dp))
                                .background(
                                    if (isSelected) Color(0xFF0284C7) else Color(0xFF1E293B).copy(alpha = 0.85f)
                                )
                                .border(
                                    1.dp,
                                    if (isSelected) Color(0xFF38BDF8) else Color(0xFF475569),
                                    RoundedCornerShape(16.dp)
                                )
                                .clickable {
                                    selectedRegion = region
                                    panOffsetX = 0f
                                    panOffsetY = 0f
                                    zoomLevel = MapZoomHelper.DEFAULT_ZOOM
                                }
                                .padding(horizontal = 12.dp, vertical = 6.dp)
                        ) {
                            Text(
                                text = region.label,
                                color = if (isSelected) Color.White else Color(0xFFCBD5E1),
                                fontSize = 11.sp,
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium
                            )
                        }
                    }
                }

                // Layer Toggles
                LayerFilterOverlay(
                    activeLayers = state.activeLayers,
                    onToggleLayer = onToggleLayer
                )

                // Proximity Warning Banner (< 800m to decomposing carcass)
                var isWarningBannerDismissed by remember { mutableStateOf(false) }
                val activeCarcassAlert = state.proximityRoadkillAlert ?: state.roadkills.firstOrNull {
                    it.clearanceStatus != ClearanceStatus.CLEARED && it.attractantThreatScore >= 0.70
                }
                if (!isWarningBannerDismissed && activeCarcassAlert != null && state.activeLayers.contains(MapLayerType.ROADKILL)) {
                    RoadkillProximityWarningBanner(
                        roadkill = activeCarcassAlert,
                        onClick = {
                            onSelectEntity(
                                SelectedEntityDetails(
                                    id = activeCarcassAlert.id,
                                    title = "Decomposing ${activeCarcassAlert.species.name.replace("_", " ")} Carcass",
                                    subtitle = "Active Apex Predator Lure • Route 4 Sector",
                                    speciesOrType = activeCarcassAlert.species.name,
                                    threatTier = AlertTier.TIER_2_WARNING,
                                    geminiBadgeText = "Gemini Forensics: ${activeCarcassAlert.decayState.name} • Threat Index ${activeCarcassAlert.attractantThreatScore}",
                                    canMarkResolved = true,
                                    isRoadkill = true,
                                    roadkillStatus = activeCarcassAlert.clearanceStatus
                                )
                            )
                        },
                        onDismiss = { isWarningBannerDismissed = true }
                    )
                }
            }

            // 3. Google Maps Floating Controls (Zoom +/- & Center Reset & Nationwide Jump)
            Column(
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(end = 16.dp, bottom = if (state.selectedEntity != null) 230.dp else 24.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                FloatingActionButton(
                    onClick = { zoomLevel = (zoomLevel + 1.0f).coerceAtMost(MapZoomHelper.MAX_ZOOM) },
                    modifier = Modifier.size(42.dp),
                    containerColor = Color(0xFF1E293B),
                    contentColor = Color.White,
                    elevation = FloatingActionButtonDefaults.elevation(4.dp)
                ) {
                    Text("+", fontSize = 20.sp, fontWeight = FontWeight.Bold)
                }

                FloatingActionButton(
                    onClick = { zoomLevel = (zoomLevel - 1.0f).coerceAtLeast(MapZoomHelper.MIN_ZOOM) },
                    modifier = Modifier.size(42.dp),
                    containerColor = Color(0xFF1E293B),
                    contentColor = Color.White,
                    elevation = FloatingActionButtonDefaults.elevation(4.dp)
                ) {
                    Text("−", fontSize = 20.sp, fontWeight = FontWeight.Bold)
                }

                // 🎯 Reset to City / District Level (Zoom 13)
                FloatingActionButton(
                    onClick = {
                        panOffsetX = 0f
                        panOffsetY = 0f
                        zoomLevel = MapZoomHelper.DEFAULT_ZOOM
                    },
                    modifier = Modifier.size(42.dp),
                    containerColor = Color(0xFF0284C7),
                    contentColor = Color.White,
                    elevation = FloatingActionButtonDefaults.elevation(4.dp)
                ) {
                    Text("🎯", fontSize = 16.sp)
                }

                // 🗾 Quick Jump to Nationwide Overview (Zoom 5)
                FloatingActionButton(
                    onClick = {
                        panOffsetX = 0f
                        panOffsetY = 0f
                        zoomLevel = MapZoomHelper.MIN_ZOOM
                    },
                    modifier = Modifier.size(42.dp),
                    containerColor = Color(0xFF1E293B),
                    contentColor = Color.White,
                    elevation = FloatingActionButtonDefaults.elevation(4.dp)
                ) {
                    Text("🗾", fontSize = 16.sp)
                }
            }

            // 4. ESRI Dark Canvas Attribution & Scale Indicator (Bottom Left)
            Row(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(start = 12.dp, bottom = if (state.selectedEntity != null) 230.dp else 12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .background(Color(0xFF0F172A).copy(alpha = 0.85f), RoundedCornerShape(4.dp))
                        .padding(horizontal = 6.dp, vertical = 3.dp)
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("ESRI Dark Canvas", fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color(0xFF38BDF8))
                        Spacer(modifier = Modifier.width(4.dp))
                        Text("• GIS", fontSize = 11.sp, fontWeight = FontWeight.Light, color = Color(0xFF94A3B8))
                        Spacer(modifier = Modifier.width(6.dp))
                        Box(modifier = Modifier.width(30.dp).height(2.dp).background(Color.White))
                        Spacer(modifier = Modifier.width(4.dp))
                        val scaleText = when {
                            zoomLevel >= 15.0f -> "100m"
                            zoomLevel >= 14.0f -> "250m"
                            zoomLevel >= 13.0f -> "1 km"
                            zoomLevel >= 12.0f -> "2 km"
                            zoomLevel >= 11.0f -> "5 km"
                            zoomLevel >= 10.0f -> "10 km"
                            zoomLevel >= 9.0f -> "25 km"
                            zoomLevel >= 8.0f -> "50 km"
                            zoomLevel >= 7.0f -> "100 km"
                            zoomLevel >= 6.0f -> "250 km"
                            else -> "500 km"
                        }
                        val zoomLabel = when {
                            zoomLevel <= 6.0f -> "Nationwide"
                            zoomLevel <= 8.0f -> "Regional"
                            zoomLevel <= 10.0f -> "Prefecture"
                            zoomLevel <= 12.0f -> "City"
                            else -> "Street"
                        }
                        Text("$scaleText ($zoomLabel z${zoomLevel.toInt()})", fontSize = 10.sp, color = Color(0xFFE2E8F0))
                    }
                }
            }

            // 5. Selected Entity Details Bottom Sheet
            state.selectedEntity?.let { entity ->
                EntityDetailCard(
                    entity = entity,
                    userRole = state.userRole,
                    onDismiss = { onSelectEntity(null) },
                    onActionClick = { onOfficialActionClick(entity.id) },
                    onRoadkillStatusChange = { newStatus -> onUpdateRoadkillStatus(entity.id, newStatus) },
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(16.dp)
                )
            }
        }
    }
}

@Composable
fun GoogleMapsVectorCanvas(
    region: MapRegion,
    state: GisMapUiState,
    panOffsetX: Float,
    panOffsetY: Float,
    zoomLevel: Float,
    onPanChange: (Float, Float) -> Unit,
    onSelectPin: (MapPinItem) -> Unit,
    onClusterClick: (ClusterDisplay) -> Unit = {}
) {
    val infiniteTransition = rememberInfiniteTransition()
    val radarPulse by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(2400),
            repeatMode = RepeatMode.Restart
        )
    )

    // Build the pins based on active region and active layers
    val pins = remember(region, state.activeLayers, state.roadkills, state.activeGeofences) {
        mutableListOf<MapPinItem>().apply {
            // Dynamic Active Geofences (AI & Manual geofences managed via Admin Console)
            if (state.activeLayers.contains(MapLayerType.GEOFENCES)) {
                state.activeGeofences.forEach { gf ->
                    val dLat = gf.centroid.latitude - region.centerLat
                    val dLon = gf.centroid.longitude - region.centerLon
                    // Show geofences within ~1500km of the current region center
                    if ((dLat * dLat + dLon * dLon) < 180.0) {
                        val badgeColor = when (gf.tier) {
                            AlertTier.TIER_3_DANGER -> TierDanger
                            AlertTier.TIER_2_WARNING -> TierWarning
                            AlertTier.TIER_1_NOTICE -> TierNotice
                            AlertTier.SAFE -> TierSafe
                        }
                        val statusLabel = when (gf.tier) {
                            AlertTier.TIER_3_DANGER -> "CRITICAL HAZARD"
                            AlertTier.TIER_2_WARNING -> "ACTIVE ALERT"
                            AlertTier.TIER_1_NOTICE -> "CAUTION ZONE"
                            AlertTier.SAFE -> "SAFE ZONE"
                        }
                        val speciesEmoji = when {
                            gf.targetSpecies.any { it.name.contains("ELEPHANT") } -> "🐘"
                            gf.targetSpecies.any { it.name.contains("TIGER") } -> "🐅"
                            gf.targetSpecies.any { it.name.contains("BEAR") } -> "🐻"
                            gf.targetSpecies.any { it.name.contains("BOAR") } -> "🐗"
                            else -> "🛡️"
                        }
                        val speciesText = gf.targetSpecies.joinToString(", ") { it.name.replace("_", " ") }.ifEmpty { "Wildlife" }
                        add(
                            MapPinItem(
                                id = gf.id,
                                lat = gf.centroid.latitude,
                                lon = gf.centroid.longitude,
                                iconEmoji = speciesEmoji,
                                title = "Geofence #${gf.id} ($statusLabel)",
                                subtitle = "${gf.causalExplanation} • Radius ${gf.radiusMeters.toInt()}m",
                                layerType = MapLayerType.GEOFENCES,
                                threatTier = gf.tier,
                                pinColor = badgeColor,
                                geminiBadge = "Autonomous AI Barrier • Target: $speciesText",
                                isGeofence = true,
                                radiusMeters = gf.radiusMeters
                            )
                        )
                    }
                }
            }

            when (region) {
                MapRegion.JAPAN_IWATE -> {
                    // Roadkill Carcasses
                    if (state.activeLayers.contains(MapLayerType.ROADKILL)) {
                        state.roadkills.forEach { rk ->
                            val (badgeColor, statusTitle) = when (rk.clearanceStatus) {
                                ClearanceStatus.REPORTED -> TierWarning to "REPORTED (Active Lure)"
                                ClearanceStatus.DISPATCHED -> Color(0xFF38BDF8) to "DISPATCHED (Sanitation En Route)"
                                ClearanceStatus.CLEARED -> TierSafe to "CLEARED (Neutralized)"
                            }
                            add(
                                MapPinItem(
                                    id = rk.id,
                                    lat = rk.location.latitude,
                                    lon = rk.location.longitude,
                                    iconEmoji = "🐗",
                                    title = "Roadkill #${rk.id} (${rk.species.name.replace("_", " ")})",
                                    subtitle = "Location: Lat ${rk.location.latitude}, Lon ${rk.location.longitude} • $statusTitle",
                                    layerType = MapLayerType.ROADKILL,
                                    threatTier = if (rk.clearanceStatus == ClearanceStatus.CLEARED) AlertTier.SAFE else AlertTier.TIER_2_WARNING,
                                    pinColor = badgeColor,
                                    geminiBadge = "Gemini Forensics: ${rk.decayState.name} • Threat Score ${rk.attractantThreatScore}",
                                    isRoadkill = true,
                                    roadkillStatus = rk.clearanceStatus
                                )
                            )
                        }
                    }

                    // Kaki Trees (Japan Scalability)
                    if (state.activeLayers.contains(MapLayerType.PERSIMMON_TREES)) {
                        add(
                            MapPinItem(
                                id = "kt_1120",
                                lat = 39.718,
                                lon = 141.162,
                                iconEmoji = "🍊",
                                title = "Top 5% Hazard Persimmon #kt_1120",
                                subtitle = "Morioka Takamatsu 3-chome • 120m to Elementary School Route",
                                layerType = MapLayerType.PERSIMMON_TREES,
                                threatTier = AlertTier.TIER_2_WARNING,
                                pinColor = Color(0xFFEA580C),
                                geminiBadge = "Gemini Vision: Abundant Fruit Load • High Fall Risk"
                            )
                        )
                        add(
                            MapPinItem(
                                id = "kt_0842",
                                lat = 39.735,
                                lon = 141.185,
                                iconEmoji = "🍊",
                                title = "Abandoned Orchard #kt_0842",
                                subtitle = "Kami-Yonai Forest Fringe • Ordinance Clearance Candidate",
                                layerType = MapLayerType.PERSIMMON_TREES,
                                threatTier = AlertTier.TIER_2_WARNING,
                                pinColor = Color(0xFFF97316),
                                geminiBadge = "Gemini Vision: Dense Crown • Pre-Hibernation Target"
                            )
                        )
                    }
                }

                MapRegion.THAILAND_KUI_BURI -> {
                    if (state.activeLayers.contains(MapLayerType.ELEPHANT_CROPS)) {
                        add(
                            MapPinItem(
                                id = "th_plot_kb_01",
                                lat = 12.062,
                                lon = 99.815,
                                iconEmoji = "🌾",
                                title = "Sugarcane Attractant Plot #KB-01",
                                subtitle = "Kui Buri Border • 14.5 ha Commercial Sugarcane",
                                layerType = MapLayerType.ELEPHANT_CROPS,
                                threatTier = AlertTier.TIER_3_DANGER,
                                pinColor = TierDanger,
                                geminiBadge = "Gemini Vision: Peak Sucrose Maturation • Downwind Scent Plume"
                            )
                        )
                        add(
                            MapPinItem(
                                id = "th_plot_kb_02",
                                lat = 12.040,
                                lon = 99.835,
                                iconEmoji = "🍌",
                                title = "Banana & Jackfruit Grove #KB-02",
                                subtitle = "Village Border Buffer • Elephant Foraging Corridor",
                                layerType = MapLayerType.ELEPHANT_CROPS,
                                threatTier = AlertTier.TIER_2_WARNING,
                                pinColor = TierWarning,
                                geminiBadge = "Gemini Vision: Ripe Odor Dispersal • Early Warning Active"
                            )
                        )
                    }
                }

                MapRegion.INDONESIA_SUMATRA -> {
                    if (state.activeLayers.contains(MapLayerType.LIVESTOCK_PREDATION)) {
                        add(
                            MapPinItem(
                                id = "ls_sumatra_01",
                                lat = 0.132,
                                lon = 101.442,
                                iconEmoji = "🐅",
                                title = "Bukit Kesuma Cattle Kill-Site (Sumatra)",
                                subtitle = "320m to Forest Edge • Active 72h Predator Caching Zone",
                                layerType = MapLayerType.LIVESTOCK_PREDATION,
                                threatTier = AlertTier.TIER_3_DANGER,
                                pinColor = TierDanger,
                                geminiBadge = "Gemini Forensics: Canine Bite Match (72mm) • Immediate Evacuation"
                            )
                        )
                        add(
                            MapPinItem(
                                id = "ls_sumatra_02",
                                lat = 0.118,
                                lon = 101.460,
                                iconEmoji = "🐄",
                                title = "Pasture Corral Perimeter #SM-04",
                                subtitle = "Palm Oil Division 3 • Vulnerable Night Herd Corrals",
                                layerType = MapLayerType.LIVESTOCK_PREDATION,
                                threatTier = AlertTier.TIER_2_WARNING,
                                pinColor = TierWarning,
                                geminiBadge = "Continuous Causal Inference: New Moon Intrusion Probability 84%"
                            )
                        )
                    }
                }
            }
        }
    }

    val coroutineScope = rememberCoroutineScope()
    val tileManager = remember { TileManager() }
    val cacheVersion = tileManager.cacheVersion // trigger recomposition on tile arrival

    BoxWithConstraints(
        modifier = Modifier
            .fillMaxSize()
            .pointerInput(Unit) {
                detectDragGestures { change, dragAmount ->
                    change.consume()
                    onPanChange(dragAmount.x, dragAmount.y)
                }
            }
    ) {
        val density = LocalDensity.current
        val screenWPx = with(density) { maxWidth.toPx() }
        val screenHPx = with(density) { maxHeight.toPx() }
        val screenCenterX = screenWPx / 2f
        val screenCenterY = screenHPx / 2f

        Canvas(modifier = Modifier.fillMaxSize()) {
            val relativeZoom = 2.0f.pow(zoomLevel - 13.0f)
            val center = Offset(size.width / 2f + panOffsetX, size.height / 2f + panOffsetY)

            // 1. Base Dark Ground
            drawRect(
                color = Color(0xFF181A1F),
                size = size
            )

            // 2. Real Slippy Map Raster Tiles (CartoDB Dark Matter / ESRI Dark Canvas)
            drawCartoDbTiles(
                region = region,
                zoomLevel = zoomLevel,
                panOffsetX = panOffsetX,
                panOffsetY = panOffsetY,
                tileManager = tileManager,
                scope = coroutineScope
            )

            // 3. Fallback vector features while initial tiles load
            if (tileManager.cacheVersion == 0) {
                drawCityBlocks(center, relativeZoom)
                drawRiversAndWaterways(center, relativeZoom, region)
                drawRoadwaysNetwork(center, relativeZoom, region)
                drawTopographyAndForest(center, relativeZoom, region)
            }

            val metrics = MapZoomHelper.calculateMetrics(zoomLevel)
            val discreteZoom = metrics.discreteZoom
            val tileSizePx = metrics.tileSizePx
            val centerTileX = TileCoordinateUtils.lonToTileX(region.centerLon, discreteZoom)
            val centerTileY = TileCoordinateUtils.latToTileY(region.centerLat, discreteZoom)
            val screenCenterX = size.width / 2f
            val screenCenterY = size.height / 2f

            val geoToOffset: (Double, Double) -> Offset = { lat, lon ->
                val tX = TileCoordinateUtils.lonToTileX(lon, discreteZoom)
                val tY = TileCoordinateUtils.latToTileY(lat, discreteZoom)
                val pxX = screenCenterX + (tX - centerTileX).toFloat() * tileSizePx + panOffsetX
                val pxY = screenCenterY + (tY - centerTileY).toFloat() * tileSizePx + panOffsetY
                Offset(pxX, pxY)
            }

            // 5. Dynamic Geofence Hazard Polygons (Live Backend Geofences & GeoJSON Simulation)
            drawGeofencePolygons(
                center = center,
                zoom = relativeZoom,
                region = region,
                geofences = state.activeGeofences,
                isLayerActive = state.activeLayers.contains(MapLayerType.GEOFENCES),
                geoToOffset = geoToOffset
            )

            // 6. Scent Plume & Wind Vector (Causal Layer)
            drawCausalScentPlumes(center, relativeZoom, region)

            // 7. User GPS Location with 800m Hazard Radar Pulse
            val userScreenPos = Offset(center.x, center.y + 10f * relativeZoom)
            drawUserLocationWithRadar(userScreenPos, radarPulse, relativeZoom)
        }

        // Overlay Interactive Clickable Pin Badges on the Screen (Slippy Tile Coordinate aligned)
        val metrics = MapZoomHelper.calculateMetrics(zoomLevel)
        val discreteZoom = metrics.discreteZoom
        val tileSizePx = metrics.tileSizePx

        val centerTileX = TileCoordinateUtils.lonToTileX(region.centerLon, discreteZoom)
        val centerTileY = TileCoordinateUtils.latToTileY(region.centerLat, discreteZoom)

        val clusterThresholdPx = 54f // Pixel proximity threshold for clustering

        // 1. Calculate screen coordinates for all active pins
        val pinPositions = pins.map { pin ->
            val pinTileX = TileCoordinateUtils.lonToTileX(pin.lon, discreteZoom)
            val pinTileY = TileCoordinateUtils.latToTileY(pin.lat, discreteZoom)
            val pinPxX = screenCenterX + (pinTileX - centerTileX).toFloat() * tileSizePx + panOffsetX
            val pinPxY = screenCenterY + (pinTileY - centerTileY).toFloat() * tileSizePx + panOffsetY
            Triple(pin, pinPxX, pinPxY)
        }

        // 2. Group close pins into clusters when zoomed out (zoomLevel <= 13.5f)
        val displayItems = mutableListOf<MapDisplayItem>()
        val visited = BooleanArray(pinPositions.size)

        for (i in pinPositions.indices) {
            if (visited[i]) continue
            visited[i] = true
            val (pinI, xI, yI) = pinPositions[i]
            val clusterGroup = mutableListOf(pinI)
            var sumX = xI
            var sumY = yI
            var sumLat = pinI.lat
            var sumLon = pinI.lon

            if (zoomLevel <= 13.5f) {
                for (j in (i + 1) until pinPositions.size) {
                    if (visited[j]) continue
                    val (pinJ, xJ, yJ) = pinPositions[j]
                    val dx = xI - xJ
                    val dy = yI - yJ
                    if (dx * dx + dy * dy <= clusterThresholdPx * clusterThresholdPx) {
                        visited[j] = true
                        clusterGroup.add(pinJ)
                        sumX += xJ
                        sumY += yJ
                        sumLat += pinJ.lat
                        sumLon += pinJ.lon
                    }
                }
            }

            if (clusterGroup.size == 1) {
                displayItems.add(SinglePinDisplay(pinI, xI, yI))
            } else {
                val count = clusterGroup.size
                val highestTier = when {
                    clusterGroup.any { it.threatTier == AlertTier.TIER_3_DANGER } -> AlertTier.TIER_3_DANGER
                    clusterGroup.any { it.threatTier == AlertTier.TIER_2_WARNING } -> AlertTier.TIER_2_WARNING
                    clusterGroup.any { it.threatTier == AlertTier.TIER_1_NOTICE } -> AlertTier.TIER_1_NOTICE
                    else -> AlertTier.SAFE
                }
                val badgeColor = when (highestTier) {
                    AlertTier.TIER_3_DANGER -> TierDanger
                    AlertTier.TIER_2_WARNING -> TierWarning
                    AlertTier.TIER_1_NOTICE -> TierNotice
                    AlertTier.SAFE -> TierSafe
                }
                displayItems.add(
                    ClusterDisplay(
                        pins = clusterGroup,
                        screenX = sumX / count,
                        screenY = sumY / count,
                        highestTier = highestTier,
                        badgeColor = badgeColor,
                        centroidLat = sumLat / count,
                        centroidLon = sumLon / count
                    )
                )
            }
        }

        // 3. Render clustered badges and individual pins
        displayItems.forEach { item ->
            // Only render items near or within visible screen area to optimize performance
            if (item.screenX >= -200f && item.screenX <= screenWPx + 200f && item.screenY >= -100f && item.screenY <= screenHPx + 100f) {
                when (item) {
                    is ClusterDisplay -> {
                        val clusterIcon = when (item.highestTier) {
                            AlertTier.TIER_3_DANGER -> "🚨"
                            AlertTier.TIER_2_WARNING -> "⚠️"
                            else -> "📍"
                        }
                        Box(
                            modifier = Modifier
                                .offset {
                                    IntOffset(
                                        (item.screenX - 22f).roundToInt(),
                                        (item.screenY - 22f).roundToInt()
                                    )
                                }
                                .clip(CircleShape)
                                .background(Color(0xFF0F172A).copy(alpha = 0.95f))
                                .border(2.5.dp, item.badgeColor, CircleShape)
                                .clickable { onClusterClick(item) }
                                .padding(horizontal = 9.dp, vertical = 6.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text(clusterIcon, fontSize = 12.sp)
                                Spacer(modifier = Modifier.width(3.dp))
                                Text(
                                    text = "${item.pins.size}",
                                    color = Color.White,
                                    fontSize = 12.sp,
                                    fontWeight = FontWeight.Black
                                )
                            }
                        }
                    }

                    is SinglePinDisplay -> {
                        val pin = item.pin
                        if (zoomLevel <= 9.0f) {
                            // Compact circular hazard emoji badge for nationwide / regional overview
                            Box(
                                modifier = Modifier
                                    .offset {
                                        IntOffset(
                                            (item.screenX - 16f).roundToInt(),
                                            (item.screenY - 16f).roundToInt()
                                        )
                                    }
                                    .size(32.dp)
                                    .clip(CircleShape)
                                    .background(Color(0xFF0F172A).copy(alpha = 0.94f))
                                    .border(2.dp, pin.pinColor, CircleShape)
                                    .clickable { onSelectPin(pin) },
                                contentAlignment = Alignment.Center
                            ) {
                                Text(pin.iconEmoji, fontSize = 16.sp)
                            }
                        } else {
                            // Rich descriptive badge for city / district / street view
                            Box(
                                modifier = Modifier
                                    .offset {
                                        IntOffset(
                                            (item.screenX - 55f).roundToInt(),
                                            (item.screenY - 16f).roundToInt()
                                        )
                                    }
                                    .clip(RoundedCornerShape(20.dp))
                                    .background(Color(0xFF0F172A).copy(alpha = 0.94f))
                                    .border(2.dp, pin.pinColor, RoundedCornerShape(20.dp))
                                    .clickable { onSelectPin(pin) }
                                    .padding(horizontal = 8.dp, vertical = 4.dp)
                            ) {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Text(pin.iconEmoji, fontSize = 14.sp)
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text(
                                        text = pin.title.take(18) + if (pin.title.length > 18) "…" else "",
                                        color = Color.White,
                                        fontSize = 10.sp,
                                        fontWeight = FontWeight.Bold
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

// Draw real CartoDB Dark Matter / OpenStreetMap raster tiles
private fun DrawScope.drawCartoDbTiles(
    region: MapRegion,
    zoomLevel: Float,
    panOffsetX: Float,
    panOffsetY: Float,
    tileManager: TileManager,
    scope: CoroutineScope
) {
    val metrics = MapZoomHelper.calculateMetrics(zoomLevel)
    val discreteZoom = metrics.discreteZoom
    val tileSizePx = metrics.tileSizePx

    val centerTileX = TileCoordinateUtils.lonToTileX(region.centerLon, discreteZoom)
    val centerTileY = TileCoordinateUtils.latToTileY(region.centerLat, discreteZoom)

    val currentCenterTileX = centerTileX - (panOffsetX / tileSizePx)
    val currentCenterTileY = centerTileY - (panOffsetY / tileSizePx)

    val screenCenterX = size.width / 2f
    val screenCenterY = size.height / 2f

    val tilesXNeeded = (screenCenterX / tileSizePx).toInt() + 2
    val tilesYNeeded = (screenCenterY / tileSizePx).toInt() + 2

    val minTileX = (currentCenterTileX - tilesXNeeded).toInt()
    val maxTileX = (currentCenterTileX + tilesXNeeded).toInt()
    val minTileY = (currentCenterTileY - tilesYNeeded).toInt()
    val maxTileY = (currentCenterTileY + tilesYNeeded).toInt()

    for (tx in minTileX..maxTileX) {
        for (ty in minTileY..maxTileY) {
            val bitmap = tileManager.getTile(discreteZoom, tx, ty, scope)
            val posX = screenCenterX + (tx - currentCenterTileX).toFloat() * tileSizePx
            val posY = screenCenterY + (ty - currentCenterTileY).toFloat() * tileSizePx

            if (bitmap != null) {
                drawImage(
                    image = bitmap,
                    dstOffset = IntOffset(posX.toInt(), posY.toInt()),
                    dstSize = IntSize(tileSizePx.toInt(), tileSizePx.toInt())
                )
            } else {
                // Subtle grid placeholder while loading tile
                drawRect(
                    color = Color(0xFF1E293B).copy(alpha = 0.4f),
                    topLeft = Offset(posX, posY),
                    size = Size(tileSizePx, tileSizePx)
                )
                drawRect(
                    color = Color(0xFF334155).copy(alpha = 0.2f),
                    topLeft = Offset(posX, posY),
                    size = Size(tileSizePx, tileSizePx),
                    style = Stroke(1f)
                )
            }
        }
    }
}

// Draw Google Maps styled city blocks
private fun DrawScope.drawCityBlocks(center: Offset, zoom: Float) {
    val blockColor = Color(0xFF20232A)
    val gridSize = 80f * zoom
    val cols = (size.width / gridSize).toInt() + 4
    val rows = (size.height / gridSize).toInt() + 4

    for (c in -cols..cols) {
        for (r in -rows..rows) {
            val left = center.x + c * gridSize + 6f
            val top = center.y + r * gridSize + 6f
            val blockW = gridSize - 12f
            val blockH = gridSize - 12f

            if (left + blockW >= 0 && left <= size.width && top + blockH >= 0 && top <= size.height) {
                drawRect(
                    color = blockColor,
                    topLeft = Offset(left, top),
                    size = Size(blockW, blockH)
                )
            }
        }
    }
}

// Draw Topographic Hillslopes & Protected Reserves (Deep Forest Green)
private fun DrawScope.drawTopographyAndForest(center: Offset, zoom: Float, region: MapRegion) {
    val forestColor = Color(0xFF132A1C)
    val reservePath = Path().apply {
        moveTo(center.x - 300f * zoom, center.y - 400f * zoom)
        cubicTo(
            center.x - 100f * zoom, center.y - 350f * zoom,
            center.x + 200f * zoom, center.y - 450f * zoom,
            center.x + 450f * zoom, center.y - 300f * zoom
        )
        lineTo(center.x + 500f * zoom, center.y - 500f * zoom)
        lineTo(center.x - 400f * zoom, center.y - 500f * zoom)
        close()
    }
    drawPath(reservePath, color = forestColor, style = Fill)

    // Secondary forest boundary
    val southForest = Path().apply {
        moveTo(center.x - 400f * zoom, center.y + 250f * zoom)
        cubicTo(
            center.x - 150f * zoom, center.y + 200f * zoom,
            center.x + 100f * zoom, center.y + 350f * zoom,
            center.x + 380f * zoom, center.y + 280f * zoom
        )
        lineTo(center.x + 400f * zoom, center.y + 500f * zoom)
        lineTo(center.x - 450f * zoom, center.y + 500f * zoom)
        close()
    }
    drawPath(southForest, color = forestColor.copy(alpha = 0.85f), style = Fill)
}

// Draw Rivers & Waterways (Kitakami River style)
private fun DrawScope.drawRiversAndWaterways(center: Offset, zoom: Float, region: MapRegion) {
    val riverColor = Color(0xFF1A365D) // Deep Map River Blue
    val riverPath = Path().apply {
        moveTo(center.x - 220f * zoom, center.y - 500f * zoom)
        cubicTo(
            center.x - 180f * zoom, center.y - 200f * zoom,
            center.x - 260f * zoom, center.y + 100f * zoom,
            center.x - 160f * zoom, center.y + 500f * zoom
        )
    }
    drawPath(riverPath, color = riverColor, style = Stroke(width = 24f * zoom))

    // Tributary Stream
    val streamPath = Path().apply {
        moveTo(center.x - 210f * zoom, center.y - 40f * zoom)
        cubicTo(
            center.x - 60f * zoom, center.y - 80f * zoom,
            center.x + 80f * zoom, center.y - 120f * zoom,
            center.x + 300f * zoom, center.y - 100f * zoom
        )
    }
    drawPath(streamPath, color = riverColor.copy(alpha = 0.8f), style = Stroke(width = 10f * zoom))
}

// Draw Roadways Network (Highways, Main Arterials, Secondary Streets)
private fun DrawScope.drawRoadwaysNetwork(center: Offset, zoom: Float, region: MapRegion) {
    // 1. National Highway (Route 4 / Asian Highway 2) - Amber / Gold Highway
    val highwayPath = Path().apply {
        moveTo(center.x - 30f * zoom, center.y - 500f * zoom)
        cubicTo(
            center.x - 10f * zoom, center.y - 150f * zoom,
            center.x - 60f * zoom, center.y + 150f * zoom,
            center.x - 20f * zoom, center.y + 500f * zoom
        )
    }
    // Highway border casing
    drawPath(highwayPath, color = Color(0xFF78350F), style = Stroke(width = 9f * zoom))
    // Highway inner core
    drawPath(highwayPath, color = Color(0xFFF59E0B), style = Stroke(width = 6f * zoom))

    // 2. Prefectural Route 396 (Detour Bypass)
    val bypassPath = Path().apply {
        moveTo(center.x + 180f * zoom, center.y - 500f * zoom)
        cubicTo(
            center.x + 140f * zoom, center.y - 100f * zoom,
            center.x + 220f * zoom, center.y + 200f * zoom,
            center.x + 160f * zoom, center.y + 500f * zoom
        )
    }
    drawPath(bypassPath, color = Color(0xFF64748B), style = Stroke(width = 5f * zoom))
    drawPath(bypassPath, color = Color(0xFF94A3B8), style = Stroke(width = 3.5f * zoom))

    // 3. East-West Connecting Avenues
    val avenue1 = Path().apply {
        moveTo(center.x - 500f * zoom, center.y - 60f * zoom)
        lineTo(center.x + 500f * zoom, center.y - 60f * zoom)
    }
    drawPath(avenue1, color = Color(0xFF334155), style = Stroke(width = 4f * zoom))

    val avenue2 = Path().apply {
        moveTo(center.x - 500f * zoom, center.y + 140f * zoom)
        lineTo(center.x + 500f * zoom, center.y + 140f * zoom)
    }
    drawPath(avenue2, color = Color(0xFF334155), style = Stroke(width = 4f * zoom))
}

// Draw Dynamic Geofence Hazard Polygons (Tier 1/2/3 GeoJSON simulation & Live Backend Geofences)
private fun DrawScope.drawGeofencePolygons(
    center: Offset,
    zoom: Float,
    region: MapRegion,
    geofences: List<GeofenceZone> = emptyList(),
    isLayerActive: Boolean = true,
    geoToOffset: (Double, Double) -> Offset = { _, _ -> Offset.Zero }
) {
    if (!isLayerActive) return

    if (geofences.isNotEmpty()) {
        // Sort so outermost buffer zones (Tier 1 Notice) render first (base layer),
        // and innermost core danger zones (Tier 3 Danger) render on top.
        val sortedGeofences = geofences.sortedBy { gf ->
            when (gf.tier) {
                AlertTier.TIER_1_NOTICE -> 1
                AlertTier.TIER_2_WARNING -> 2
                AlertTier.TIER_3_DANGER -> 3
                AlertTier.SAFE -> 0
            }
        }

        sortedGeofences.forEach { gf ->
            val zoneColor = when (gf.tier) {
                AlertTier.TIER_3_DANGER -> TierDanger
                AlertTier.TIER_2_WARNING -> TierWarning
                AlertTier.TIER_1_NOTICE -> TierNotice
                AlertTier.SAFE -> TierSafe
            }
            val fillAlpha = when (gf.tier) {
                AlertTier.TIER_3_DANGER -> 0.28f
                AlertTier.TIER_2_WARNING -> 0.18f
                AlertTier.TIER_1_NOTICE -> 0.10f
                AlertTier.SAFE -> 0.05f
            }

            if (gf.coordinates.size >= 3) {
                val path = Path().apply {
                    val first = gf.coordinates[0]
                    val firstPt = geoToOffset(first.latitude, first.longitude)
                    moveTo(firstPt.x, firstPt.y)
                    for (i in 1 until gf.coordinates.size) {
                        val pt = gf.coordinates[i]
                        val off = geoToOffset(pt.latitude, pt.longitude)
                        lineTo(off.x, off.y)
                    }
                    close()
                }
                drawPath(path, color = zoneColor.copy(alpha = fillAlpha), style = Fill)
                drawPath(
                    path,
                    color = zoneColor,
                    style = Stroke(
                        width = if (gf.tier == AlertTier.TIER_3_DANGER) (2.5f * zoom).coerceIn(2.0f, 4.5f) else (1.8f * zoom).coerceIn(1.2f, 3.0f),
                        pathEffect = when (gf.tier) {
                            AlertTier.TIER_2_WARNING -> PathEffect.dashPathEffect(floatArrayOf(12f, 8f), 0f)
                            AlertTier.TIER_1_NOTICE -> PathEffect.dashPathEffect(floatArrayOf(16f, 12f), 0f)
                            else -> null
                        }
                    )
                )
            } else if (gf.centroid.latitude != 0.0 || gf.centroid.longitude != 0.0) {
                val centerPos = geoToOffset(gf.centroid.latitude, gf.centroid.longitude)
                val edgePos = geoToOffset(
                    gf.centroid.latitude,
                    gf.centroid.longitude + (gf.radiusMeters / 111320.0)
                )
                val rPx = kotlin.math.hypot(edgePos.x - centerPos.x, edgePos.y - centerPos.y).coerceAtLeast(12f)
                drawCircle(
                    color = zoneColor.copy(alpha = fillAlpha),
                    radius = rPx,
                    center = centerPos,
                    style = Fill
                )
                drawCircle(
                    color = zoneColor,
                    radius = rPx,
                    center = centerPos,
                    style = Stroke(
                        width = if (gf.tier == AlertTier.TIER_3_DANGER) (2.5f * zoom).coerceIn(2.0f, 4.5f) else (1.8f * zoom).coerceIn(1.2f, 3.0f),
                        pathEffect = when (gf.tier) {
                            AlertTier.TIER_2_WARNING -> PathEffect.dashPathEffect(floatArrayOf(10f, 6f), 0f)
                            AlertTier.TIER_1_NOTICE -> PathEffect.dashPathEffect(floatArrayOf(14f, 10f), 0f)
                            else -> null
                        }
                    )
                )
            }
        }
        return
    }

    when (region) {
        MapRegion.JAPAN_IWATE -> {
            // Tier 2 Warning Bear Corridor Polygon (Orange)
            val bearCorridor = Path().apply {
                moveTo(center.x + 20f * zoom, center.y - 280f * zoom)
                lineTo(center.x + 260f * zoom, center.y - 240f * zoom)
                lineTo(center.x + 220f * zoom, center.y - 60f * zoom)
                lineTo(center.x + 60f * zoom, center.y - 90f * zoom)
                close()
            }
            drawPath(bearCorridor, color = Color(0xFFF97316).copy(alpha = 0.22f), style = Fill)
            drawPath(
                bearCorridor,
                color = Color(0xFFF97316),
                style = Stroke(
                    width = 2.5f * zoom,
                    pathEffect = PathEffect.dashPathEffect(floatArrayOf(12f, 8f), 0f)
                )
            )

            // Tier 1 Notice Zone (Yellow)
            val noticeZone = Path().apply {
                moveTo(center.x - 300f * zoom, center.y + 80f * zoom)
                lineTo(center.x - 120f * zoom, center.y + 100f * zoom)
                lineTo(center.x - 150f * zoom, center.y + 240f * zoom)
                lineTo(center.x - 320f * zoom, center.y + 210f * zoom)
                close()
            }
            drawPath(noticeZone, color = Color(0xFFEAB308).copy(alpha = 0.18f), style = Fill)
            drawPath(
                noticeZone,
                color = Color(0xFFEAB308),
                style = Stroke(width = 1.8f * zoom)
            )
        }

        MapRegion.THAILAND_KUI_BURI -> {
            // Tier 3 Danger Elephant Crop Raid Polygon (Red)
            val elephantHazard = Path().apply {
                moveTo(center.x - 80f * zoom, center.y - 120f * zoom)
                lineTo(center.x + 180f * zoom, center.y - 150f * zoom)
                lineTo(center.x + 220f * zoom, center.y + 60f * zoom)
                lineTo(center.x - 40f * zoom, center.y + 90f * zoom)
                close()
            }
            drawPath(elephantHazard, color = Color(0xFFEF4444).copy(alpha = 0.25f), style = Fill)
            drawPath(
                elephantHazard,
                color = Color(0xFFEF4444),
                style = Stroke(width = 2.5f * zoom)
            )
        }

        MapRegion.INDONESIA_SUMATRA -> {
            // Tier 3 Danger Tiger Kill-Site 72h Caching Circle (Red Pulse)
            drawCircle(
                color = Color(0xFFEF4444).copy(alpha = 0.22f),
                radius = 110f * zoom,
                center = Offset(center.x + 60f * zoom, center.y - 80f * zoom)
            )
            drawCircle(
                color = Color(0xFFEF4444),
                radius = 110f * zoom,
                center = Offset(center.x + 60f * zoom, center.y - 80f * zoom),
                style = Stroke(
                    width = 2.5f * zoom,
                    pathEffect = PathEffect.dashPathEffect(floatArrayOf(10f, 6f), 0f)
                )
            )
        }
    }
}

// Draw Scent Plumes & Wind Vector (Olfactory Causal Modeling)
private fun DrawScope.drawCausalScentPlumes(center: Offset, zoom: Float, region: MapRegion) {
    val windColor = Color(0xFF38BDF8).copy(alpha = 0.35f)
    for (i in 0..4) {
        val yOffset = center.y - 200f * zoom + (i * 35f * zoom)
        val startX = center.x + 240f * zoom
        val endX = center.x + 40f * zoom

        val path = Path().apply {
            moveTo(startX, yOffset)
            quadraticBezierTo(
                startX - 100f * zoom, yOffset + 15f * zoom,
                endX, yOffset + (sin(i.toDouble()) * 20f * zoom).toFloat()
            )
        }
        drawPath(
            path,
            color = windColor,
            style = Stroke(
                width = 2f * zoom,
                pathEffect = PathEffect.dashPathEffect(floatArrayOf(8f, 6f), 0f)
            )
        )
    }
}

// Draw User Location with Radar Sensor Pulse Ring
private fun DrawScope.drawUserLocationWithRadar(pos: Offset, pulseProgress: Float, zoom: Float) {
    val pulseRadius = (40f + pulseProgress * 120f) * zoom
    val pulseAlpha = (1f - pulseProgress).coerceIn(0f, 1f) * 0.45f

    // Outer Expanding Radar Pulse
    drawCircle(
        color = Color(0xFF38BDF8).copy(alpha = pulseAlpha),
        radius = pulseRadius,
        center = pos
    )
    drawCircle(
        color = Color(0xFF38BDF8).copy(alpha = pulseAlpha * 0.8f),
        radius = pulseRadius,
        center = pos,
        style = Stroke(width = 2f)
    )

    // Inner GPS Location Indicator
    drawCircle(
        color = Color(0xFF0284C7).copy(alpha = 0.3f),
        radius = 16f * zoom,
        center = pos
    )
    drawCircle(
        color = Color.White,
        radius = 8f * zoom,
        center = pos
    )
    drawCircle(
        color = Color(0xFF0284C7),
        radius = 6f * zoom,
        center = pos
    )
}

@Composable
fun RoadkillProximityWarningBanner(
    roadkill: RoadkillRecord,
    onClick: () -> Unit,
    onDismiss: () -> Unit = {}
) {
    val infiniteTransition = rememberInfiniteTransition()
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.85f,
        targetValue = 1.0f,
        animationSpec = infiniteRepeatable(
            animation = tween(500),
            repeatMode = RepeatMode.Reverse
        )
    )

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.5.dp, TierDanger.copy(alpha = pulseAlpha), RoundedCornerShape(10.dp))
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(10.dp),
        colors = CardDefaults.cardColors(containerColor = TierDanger.copy(alpha = 0.95f))
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("⚠️", fontSize = 16.sp)
            Spacer(modifier = Modifier.width(8.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "ROADKILL ALERT (< 800m) • ${roadkill.species.name.replace("_", " ")} (${roadkill.decayState.name})",
                    color = Color.White,
                    fontWeight = FontWeight.Bold,
                    fontSize = 11.sp
                )
                Text(
                    text = "Threat Index: ${roadkill.attractantThreatScore} • High apex predator lure. Do not stop.",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 10.sp
                )
            }
            Spacer(modifier = Modifier.width(6.dp))
            Box(
                modifier = Modifier
                    .background(Color.White, RoundedCornerShape(4.dp))
                    .padding(horizontal = 6.dp, vertical = 3.dp)
            ) {
                Text("INSPECT", fontSize = 9.sp, fontWeight = FontWeight.Bold, color = TierDanger)
            }
            Spacer(modifier = Modifier.width(6.dp))
            Box(
                modifier = Modifier
                    .clip(CircleShape)
                    .clickable { onDismiss() }
                    .padding(4.dp)
            ) {
                Text("✕", fontSize = 11.sp, color = Color.White.copy(alpha = 0.8f))
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LayerFilterOverlay(
    activeLayers: Set<MapLayerType>,
    onToggleLayer: (MapLayerType) -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .horizontalScroll(rememberScrollState()),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        MapLayerType.values().forEach { layer ->
            val isSelected = activeLayers.contains(layer)
            FilterChip(
                selected = isSelected,
                onClick = { onToggleLayer(layer) },
                label = { Text(layer.label, fontSize = 11.sp) },
                shape = RoundedCornerShape(20.dp),
                colors = FilterChipDefaults.filterChipColors(
                    selectedContainerColor = Color(0xFF0284C7),
                    selectedLabelColor = Color.White,
                    containerColor = Color(0xFF1E293B).copy(alpha = 0.85f),
                    labelColor = Color(0xFFCBD5E1)
                )
            )
        }
    }
}

@Composable
fun EntityDetailCard(
    entity: SelectedEntityDetails,
    userRole: UserRole,
    onDismiss: () -> Unit,
    onActionClick: () -> Unit,
    onRoadkillStatusChange: (ClearanceStatus) -> Unit = {},
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = Color(0xFF0F172A).copy(alpha = 0.96f)
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
    ) {
        Column(
            modifier = Modifier.padding(18.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = entity.title,
                    fontWeight = FontWeight.Bold,
                    fontSize = 15.sp,
                    color = Color.White
                )
                Box(
                    modifier = Modifier
                        .clickable(onClick = onDismiss)
                        .padding(4.dp)
                ) {
                    Text("✕", fontSize = 14.sp, color = Color(0xFF94A3B8))
                }
            }

            Text(
                text = entity.subtitle,
                fontSize = 12.sp,
                color = Color(0xFF94A3B8)
            )

            // Status chip for roadkill
            if (entity.isRoadkill && entity.roadkillStatus != null) {
                val (chipColor, label) = when (entity.roadkillStatus) {
                    ClearanceStatus.REPORTED -> TierWarning to "STATUS: REPORTED (Pending Sanitation Dispatch)"
                    ClearanceStatus.DISPATCHED -> Color(0xFF38BDF8) to "STATUS: DISPATCHED (Truck En Route)"
                    ClearanceStatus.CLEARED -> TierSafe to "STATUS: CLEARED (Mitigation Complete)"
                }
                Box(
                    modifier = Modifier
                        .background(chipColor.copy(alpha = 0.2f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = label,
                        color = chipColor,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            // Status chip for geofence barrier
            if (entity.isGeofence) {
                val (chipColor, label) = when (entity.threatTier) {
                    AlertTier.TIER_3_DANGER -> TierDanger to "DEFENSIVE PERIMETER: CRITICAL (Detour Required)"
                    AlertTier.TIER_2_WARNING -> TierWarning to "DEFENSIVE PERIMETER: ACTIVE WARNING (Proceed with Caution)"
                    AlertTier.TIER_1_NOTICE -> TierNotice to "DEFENSIVE PERIMETER: CAUTION ZONE"
                    AlertTier.SAFE -> TierSafe to "DEFENSIVE PERIMETER: SAFE BUFFER"
                }
                Box(
                    modifier = Modifier
                        .background(chipColor.copy(alpha = 0.2f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = label,
                        color = chipColor,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            entity.geminiBadgeText?.let { badge ->
                Box(
                    modifier = Modifier
                        .background(Color(0xFF38BDF8).copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = badge,
                        color = Color(0xFF38BDF8),
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium
                    )
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            // Action row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedButton(
                    onClick = onDismiss,
                    modifier = Modifier.weight(0.9f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text("Close", fontSize = 12.sp, color = Color.White)
                }

                // Quick jump to real Google Maps external navigation & street view
                val uriHandler = LocalUriHandler.current
                if (entity.latitude != null && entity.longitude != null) {
                    OutlinedButton(
                        onClick = {
                            val url = "https://www.google.com/maps/search/?api=1&query=${entity.latitude},${entity.longitude}"
                            uriHandler.openUri(url)
                        },
                        modifier = Modifier.weight(1.3f),
                        shape = RoundedCornerShape(8.dp),
                        border = BorderStroke(1.dp, Color(0xFF38BDF8))
                    ) {
                        Text("🗺️ Google Maps", fontSize = 11.sp, color = Color(0xFF38BDF8), fontWeight = FontWeight.Bold)
                    }
                }

                // If roadkill progression
                if (entity.isRoadkill) {
                    when (entity.roadkillStatus) {
                        ClearanceStatus.REPORTED -> {
                            Button(
                                onClick = { onRoadkillStatusChange(ClearanceStatus.DISPATCHED) },
                                modifier = Modifier.weight(1.8f),
                                shape = RoundedCornerShape(8.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF0284C7))
                            ) {
                                Text("Dispatch Sanitation Truck", fontSize = 11.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                        ClearanceStatus.DISPATCHED -> {
                            Button(
                                onClick = { onRoadkillStatusChange(ClearanceStatus.CLEARED) },
                                modifier = Modifier.weight(1.8f),
                                shape = RoundedCornerShape(8.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = TierSafe)
                            ) {
                                Text("Confirm Carcass Cleared", fontSize = 11.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                        ClearanceStatus.CLEARED, null -> {
                            Box(
                                modifier = Modifier
                                    .weight(1.5f)
                                    .background(TierSafe.copy(alpha = 0.2f), RoundedCornerShape(8.dp))
                                    .padding(vertical = 10.dp),
                                contentAlignment = Alignment.Center
                            ) {
                                Text("Carcass Cleared ✓", color = TierSafe, fontWeight = FontWeight.Bold, fontSize = 12.sp)
                            }
                        }
                    }
                } else if ((userRole == UserRole.OFFICIAL || userRole == UserRole.ADMIN) && entity.canMarkResolved) {
                    Button(
                        onClick = onActionClick,
                        modifier = Modifier.weight(1.5f),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text("Official: Mark Resolved", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}
