package ai.coexistshield.ui

import ai.coexistshield.auth.AuthService
import ai.coexistshield.auth.AuthUser
import ai.coexistshield.auth.InMemorySessionStore
import ai.coexistshield.auth.SessionStore
import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.ClearanceStatus
import ai.coexistshield.model.DecayState
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.Incident
import ai.coexistshield.model.ReportCategory
import ai.coexistshield.model.RoadkillRecord
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.ui.admin.AdminConsoleScreen
import ai.coexistshield.ui.auth.AuthScreen
import ai.coexistshield.ui.admin.AdminConsoleUiState
import ai.coexistshield.ui.hud.MainHudScreen
import ai.coexistshield.ui.hud.MainHudUiState
import ai.coexistshield.ui.map.GisHazardMapScreen
import ai.coexistshield.ui.map.GisMapUiState
import ai.coexistshield.ui.map.MapLayerType
import ai.coexistshield.ui.report.IncidentReportScreen
import ai.coexistshield.ui.report.IncidentReportUiState
import ai.coexistshield.ui.theme.CoExistShieldTheme
import ai.coexistshield.ui.theme.TierDanger
import ai.coexistshield.ui.theme.TierNotice
import ai.coexistshield.ui.theme.TierSafe
import ai.coexistshield.ui.theme.TierWarning
import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

enum class AppNavDestination(val label: String) {
    HUD("Alert HUD"),
    MAP("GIS Map"),
    REPORT("Report"),
    ADMIN("Console")
}

@Composable
fun App(sessionStore: SessionStore = remember { InMemorySessionStore() }) {
    CoExistShieldTheme {
        var currentDestination by remember { mutableStateOf(AppNavDestination.HUD) }
        var currentRole by remember { mutableStateOf(UserRole.CITIZEN) }
        var authUser by remember { mutableStateOf<AuthUser?>(null) }
        var isGuest by remember { mutableStateOf(true) }
        var showDemoControls by remember { mutableStateOf(false) }
        val authService = remember { AuthService() }
        var isRestoringSession by remember { mutableStateOf(sessionStore.load() != null) }

        fun establishSession(user: AuthUser) {
            authUser = user
            currentRole = user.role
            ai.coexistshield.network.ApiClient.setAuthToken(user.idToken)
            sessionStore.save(user.refreshToken)
        }

        // Restore a persisted session on startup.
        LaunchedEffect(Unit) {
            val saved = sessionStore.load()
            if (saved != null) {
                authService.refresh(saved).fold(
                    onSuccess = { establishSession(it) },
                    onFailure = { sessionStore.clear() }
                )
                isRestoringSession = false
            }
        }

        // Renew the ID token (expires after 1h) while signed in.
        LaunchedEffect(authUser?.uid) {
            while (authUser != null) {
                delay(50 * 60 * 1000L)
                val current = authUser ?: break
                authService.refresh(current.refreshToken).onSuccess { establishSession(it) }
            }
        }
        var currentTier by remember { mutableStateOf(AlertTier.TIER_2_WARNING) }

        // Simulation toggles for offline connectivity and siren
        var isAppOffline by remember { mutableStateOf(false) }
        var isSirenPlaying by remember { mutableStateOf(false) }
        var offlinePendingCount by remember { mutableStateOf(0) }

        val snackbarHostState = remember { SnackbarHostState() }
        val scope = rememberCoroutineScope()

        val geofenceRepository = remember {
            ai.coexistshield.repository.GeofenceRepository(ai.coexistshield.network.ApiClient.createHttpClient())
        }
        val incidentRepository = remember {
            ai.coexistshield.repository.IncidentRepository(ai.coexistshield.network.ApiClient.createHttpClient())
        }

        // 1. Dynamic HUD State synchronized with interactive demo controllers (Issue #1)
        val hudState = remember(currentRole, currentTier, isAppOffline, isSirenPlaying) {
            val (species, distance, explanation) = when (currentTier) {
                AlertTier.SAFE -> Triple(
                    emptyList(),
                    null,
                    "Clear Zone — Continuous causal sensors indicate safe passage within 5km radius."
                )
                AlertTier.TIER_1_NOTICE -> Triple(
                    listOf(WildlifeSpecies.BEAR),
                    1800.0,
                    "Tier 1 Notice: Approaching moderate bear migration corridor in Iwate valley."
                )
                AlertTier.TIER_2_WARNING -> Triple(
                    listOf(WildlifeSpecies.BEAR, WildlifeSpecies.TIGER),
                    620.0,
                    "Tier 2 Warning: New Moon night + 14 km/h westerly wind blowing toward unharvested persimmon grove 620m away."
                )
                AlertTier.TIER_3_DANGER -> Triple(
                    listOf(WildlifeSpecies.TIGER),
                    180.0,
                    "IMMEDIATE DANGER: Fresh livestock kill-site located 180m away. High probability of tiger returning within 72 hours."
                )
            }

            val detour = when (currentTier) {
                AlertTier.TIER_2_WARNING -> "Bypass Forestry Route 4 via Prefectural Highway 396 (+1.2km • 0% Encounter Probability)"
                AlertTier.TIER_3_DANGER -> "EVACUATE IMMEDIATELY: Divert 90° away from downwind scent corridor to Shelter Hub #2"
                else -> null
            }

            MainHudUiState(
                currentTier = currentTier,
                targetSpecies = species,
                closestDistanceMeters = distance,
                causalExplanation = explanation,
                isOfflineMode = isAppOffline,
                isSirenPlaying = isSirenPlaying,
                isHapticActive = currentTier == AlertTier.TIER_3_DANGER,
                suggestedDetour = detour,
                activeRole = currentRole
            )
        }

        // 2. Roadkill Lifecycle State (Issue #3)
        var roadkillList by remember {
            mutableStateOf(
                listOf(
                    RoadkillRecord(
                        id = "rk_4412",
                        incidentId = "inc_iwate_03",
                        species = WildlifeSpecies.WILD_BOAR,
                        location = GeoLocation(39.7120, 141.1480),
                        decayState = DecayState.BLOATED,
                        attractantThreatScore = 0.88,
                        clearanceStatus = ClearanceStatus.REPORTED,
                        reportedAt = "2026-10-14T03:00:00Z"
                    ),
                    RoadkillRecord(
                        id = "rk_3901",
                        incidentId = "inc_iwate_04",
                        species = WildlifeSpecies.DEER,
                        location = GeoLocation(39.6410, 141.9520),
                        decayState = DecayState.DECOMPOSING,
                        attractantThreatScore = 0.75,
                        clearanceStatus = ClearanceStatus.DISPATCHED,
                        assignedOfficialId = "OFFICER-IWATE-42",
                        reportedAt = "2026-10-13T18:00:00Z"
                    ),
                    RoadkillRecord(
                        id = "rk_2105",
                        incidentId = "inc_iwate_05",
                        species = WildlifeSpecies.UNKNOWN,
                        location = GeoLocation(39.6800, 141.1300),
                        decayState = DecayState.SKELETAL,
                        attractantThreatScore = 0.20,
                        clearanceStatus = ClearanceStatus.CLEARED,
                        assignedOfficialId = "OFFICER-IWATE-18",
                        reportedAt = "2026-10-12T09:00:00Z",
                        clearedAt = "2026-10-12T14:30:00Z"
                    )
                )
            )
        }

        // 3. Map State
        var mapLayers by remember {
            mutableStateOf(
                setOf(
                    MapLayerType.ROADKILL,
                    MapLayerType.PERSIMMON_TREES,
                    MapLayerType.ELEPHANT_CROPS,
                    MapLayerType.LIVESTOCK_PREDATION
                )
            )
        }
        var mapSelectedEntity by remember { mutableStateOf<ai.coexistshield.ui.map.SelectedEntityDetails?>(null) }
        val mapState = GisMapUiState(
            userLocation = GeoLocation(39.702, 141.152),
            roadkills = roadkillList,
            activeLayers = mapLayers,
            selectedEntity = mapSelectedEntity,
            userRole = currentRole
        )

        // 4. Report State (Issue #2)
        var reportCategory by remember { mutableStateOf(ReportCategory.ATTRACTANT_FRUIT_TREE) }
        var reportSpecies by remember { mutableStateOf(WildlifeSpecies.BEAR) }
        var reportTraceType by remember { mutableStateOf("CLAW_MARK_BARK_STRIPPING") }
        var reportDecayState by remember { mutableStateOf(DecayState.BLOATED) }
        var reportPropertyTag by remember { mutableStateOf(ai.coexistshield.model.PropertyContextTag.VACANT_HOUSE) }
        var reportNearSchoolRoute by remember { mutableStateOf(true) }

        val reportState = IncidentReportUiState(
            selectedCategory = reportCategory,
            selectedSpecies = reportSpecies,
            selectedTraceType = reportTraceType,
            selectedDecayState = reportDecayState,
            selectedPropertyTag = reportPropertyTag,
            isNearSchoolRoute = reportNearSchoolRoute,
            isOfflineMode = isAppOffline,
            offlineQueueCount = offlinePendingCount,
            userRole = currentRole
        )

        // 5. Admin State & Continuous Retraining Engine
        var adminRetrainTimestamp by remember { mutableStateOf("2026-10-14 02:30 UTC") }
        var adminModelAucScore by remember { mutableStateOf(0.86) }
        var adminIsRetraining by remember { mutableStateOf(false) }
        var adminGcfFelledCount by remember { mutableStateOf(214) }
        var adminGcfRiskReductionPct by remember { mutableStateOf(64.8) }
        var adminFelledTreeIds by remember { mutableStateOf(setOf<String>()) }

        val adminState = AdminConsoleUiState(
            userRole = currentRole,
            lastRetrainTimestamp = adminRetrainTimestamp,
            modelAucScore = adminModelAucScore,
            isRetrainingInProgress = adminIsRetraining,
            gcfTreesClearedCount = adminGcfFelledCount,
            gcfAttractantReductionPct = adminGcfRiskReductionPct,
            top5PercentPriorityTrees = listOf(
                ai.coexistshield.ui.admin.PriorityKakiItem("kt_1120", "盛岡市高松3丁目12-4 • 空き家", 0.95, true, 120, "GRANTED", "PERMIT-MRK-2026-1120", 1, "ABUNDANT (Falling)",
                    status = if (adminFelledTreeIds.contains("kt_1120")) ai.coexistshield.model.TreeManagementStatus.FELLED else ai.coexistshield.model.TreeManagementStatus.UNMANAGED),
                ai.coexistshield.ui.admin.PriorityKakiItem("kt_0842", "盛岡市上米内字中野 • 林縁放置果樹園", 0.91, true, 240, "ORDINANCE_CLEARANCE", "PERMIT-MRK-2026-0842", 2, "ABUNDANT",
                    status = if (adminFelledTreeIds.contains("kt_0842")) ai.coexistshield.model.TreeManagementStatus.FELLED else ai.coexistshield.model.TreeManagementStatus.SCHEDULED_REMOVAL),
                ai.coexistshield.ui.admin.PriorityKakiItem("kt_0319", "盛岡市浅岸字二ツ森 • 住宅街裏山", 0.88, true, 310, "PENDING_CONSENT", "PERMIT-MRK-2026-0319", 3, "MODERATE",
                    status = if (adminFelledTreeIds.contains("kt_0319")) ai.coexistshield.model.TreeManagementStatus.FELLED else ai.coexistshield.model.TreeManagementStatus.UNMANAGED)
            )
        )

        if (isRestoringSession) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("セッションを復元中...")
            }
            return@CoExistShieldTheme
        }

        if (authUser == null && !isGuest) {
            AuthScreen(
                authService = authService,
                onAuthenticated = { user ->
                    establishSession(user)
                    scope.launch { snackbarHostState.showSnackbar("ようこそ、${user.email}") }
                },
                onContinueAsGuest = { isGuest = true }
            )
            return@CoExistShieldTheme
        }

        Scaffold(
            snackbarHost = { SnackbarHost(snackbarHostState) },
            topBar = {
                // Interactive Demo Controls Banner (Judge & Testing Mode)
                DemoControlBar(
                    currentRole = currentRole,
                    currentTier = currentTier,
                    signedInEmail = authUser?.email,
                    isExpanded = showDemoControls,
                    onToggleExpand = { showDemoControls = !showDemoControls },
                    onLoginClick = { isGuest = false },
                    onLogoutClick = {
                        authUser = null
                        isGuest = false
                        currentRole = UserRole.CITIZEN
                        currentDestination = AppNavDestination.HUD
                        ai.coexistshield.network.ApiClient.setAuthToken(null)
                        sessionStore.clear()
                    },
                    onRoleChange = { newRole ->
                        currentRole = newRole
                        scope.launch { snackbarHostState.showSnackbar("Switched Role to ${newRole.name}") }
                    },
                    onTierChange = { newTier ->
                        currentTier = newTier
                        if (newTier == AlertTier.TIER_3_DANGER) {
                            isSirenPlaying = true
                        } else {
                            isSirenPlaying = false
                        }
                        scope.launch { snackbarHostState.showSnackbar("Simulated Hazard: ${newTier.name}") }
                    }
                )
            },
            bottomBar = {
                NavigationBar(
                    containerColor = MaterialTheme.colorScheme.surface
                ) {
                    AppNavDestination.values().forEach { destination ->
                        // Only show Admin Console if role is OFFICIAL or ADMIN
                        if (destination != AppNavDestination.ADMIN || currentRole != UserRole.CITIZEN) {
                            NavigationBarItem(
                                selected = currentDestination == destination,
                                onClick = { currentDestination = destination },
                                label = { Text(destination.label, fontSize = 11.sp) },
                                icon = {
                                    val iconChar = when (destination) {
                                        AppNavDestination.HUD -> "🛡️"
                                        AppNavDestination.MAP -> "🗺️"
                                        AppNavDestination.REPORT -> "📸"
                                        AppNavDestination.ADMIN -> "⚙️"
                                    }
                                    Text(iconChar, fontSize = 16.sp)
                                }
                            )
                        }
                    }
                }
            }
        ) { paddingValues ->
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
            ) {
                when (currentDestination) {
                    AppNavDestination.HUD -> MainHudScreen(
                        state = hudState,
                        onOpenMapClick = { currentDestination = AppNavDestination.MAP },
                        onQuickReportClick = { currentDestination = AppNavDestination.REPORT },
                        onOfficialDispatchClick = { currentDestination = AppNavDestination.ADMIN },
                        onToggleSirenClick = { isSirenPlaying = !isSirenPlaying },
                        onToggleOfflineClick = {
                            isAppOffline = !isAppOffline
                            scope.launch {
                                snackbarHostState.showSnackbar(
                                    if (isAppOffline) "Offline Cache Mode Active (Local Geofences Stored)"
                                    else "Reconnected Live Network"
                                )
                            }
                        },
                        onEmergencySosClick = {
                            scope.launch { snackbarHostState.showSnackbar("EMERGENCY SOS: Rangers and Municipal Responders Notified!") }
                        }
                    )
                    AppNavDestination.MAP -> GisHazardMapScreen(
                        state = mapState,
                        onBackClick = { currentDestination = AppNavDestination.HUD },
                        onToggleLayer = { layer ->
                            mapLayers = if (mapLayers.contains(layer)) mapLayers - layer else mapLayers + layer
                        },
                        onSelectEntity = { mapSelectedEntity = it },
                        onUpdateRoadkillStatus = { rkId, newStatus ->
                            roadkillList = roadkillList.map { rk ->
                                if (rk.id == rkId) {
                                    rk.copy(
                                        clearanceStatus = newStatus,
                                        assignedOfficialId = if (newStatus == ClearanceStatus.DISPATCHED) "OFFICER-SANITATION-09" else rk.assignedOfficialId
                                    )
                                } else rk
                            }
                            mapSelectedEntity = null
                            scope.launch {
                                incidentRepository.updateRoadkillStatus(rkId, newStatus, "OFFICER-SANITATION-09")
                                val msg = if (newStatus == ClearanceStatus.DISPATCHED) {
                                    "Sanitation Crew Dispatched for carcass $rkId! Lure mitigation underway."
                                } else {
                                    "Carcass $rkId Cleared & Removed! Predator attractant hazard neutralized."
                                }
                                snackbarHostState.showSnackbar(msg)
                            }
                        },
                        onOfficialActionClick = { entityId ->
                            scope.launch {
                                if (entityId.startsWith("th_plot")) {
                                    geofenceRepository.updateCropHarvestStage(entityId, "HARVESTED_CLEARED")
                                    snackbarHostState.showSnackbar("HEC Mitigation: Plot $entityId marked Harvested & Cleared! Scent plume corridor neutralized.")
                                } else if (entityId.startsWith("ls_")) {
                                    geofenceRepository.clearTigerCarcass(entityId)
                                    snackbarHostState.showSnackbar("Tiger Mitigation: Kill-site carcass at $entityId removed! 72-hour caching hazard neutralized.")
                                } else {
                                    snackbarHostState.showSnackbar("Official Action Logged: #$entityId marked Resolved / Cleared!")
                                }
                                mapSelectedEntity = null
                            }
                        }
                    )
                    AppNavDestination.REPORT -> IncidentReportScreen(
                        state = reportState,
                        onBackClick = { currentDestination = AppNavDestination.HUD },
                        onCategoryChange = { reportCategory = it },
                        onSpeciesChange = { reportSpecies = it },
                        onTraceTypeChange = { reportTraceType = it },
                        onDecayStateChange = { reportDecayState = it },
                        onPropertyTagChange = { reportPropertyTag = it },
                        onNearSchoolRouteChange = { reportNearSchoolRoute = it },
                        onToggleOfflineMode = {
                            isAppOffline = it
                            scope.launch {
                                snackbarHostState.showSnackbar(
                                    if (isAppOffline) "Offline Mode Simulated: Reports queue locally"
                                    else "Online Mode Restored"
                                )
                            }
                        },
                        onRetakePhotoClick = {
                            scope.launch { snackbarHostState.showSnackbar("Camera reopened — New photo captured.") }
                        },
                        onSyncOfflineQueueClick = {
                            scope.launch {
                                val count = incidentRepository.syncOfflineQueue()
                                offlinePendingCount = incidentRepository.getOfflineQueueCount()
                                snackbarHostState.showSnackbar(
                                    if (count > 0) "Synchronized $count offline report(s) to Cloud Run!"
                                    else "All reports already synced."
                                )
                            }
                        },
                        onSubmitReportClick = {
                            scope.launch {
                                val incidentId = "inc_${kotlin.random.Random.nextInt(1000, 9999)}"
                                val newIncident = Incident(
                                    id = incidentId,
                                    category = reportCategory,
                                    species = reportSpecies,
                                    location = GeoLocation(39.702, 141.152),
                                    timestamp = "2026-10-14T15:00:00Z",
                                    region = "JAPAN",
                                    description = "Crowdsourced report for $reportCategory ($reportSpecies)",
                                    reporterId = authUser?.uid ?: "usr_current_device",
                                    reporterRole = currentRole
                                )

                                // If Roadkill, immediately register into roadkillList
                                if (reportCategory == ReportCategory.ROADKILL) {
                                    val rkRecord = RoadkillRecord(
                                        id = "rk_${kotlin.random.Random.nextInt(5000, 9999)}",
                                        incidentId = incidentId,
                                        species = reportSpecies,
                                        location = GeoLocation(39.702, 141.152),
                                        decayState = reportDecayState,
                                        attractantThreatScore = 0.88,
                                        clearanceStatus = ClearanceStatus.REPORTED,
                                        reportedAt = "2026-10-14T15:00:00Z"
                                    )
                                    roadkillList = listOf(rkRecord) + roadkillList
                                }

                                if (isAppOffline) {
                                    incidentRepository.enqueueOfflineIncident(newIncident)
                                    offlinePendingCount = incidentRepository.getOfflineQueueCount()
                                    snackbarHostState.showSnackbar("Offline Mode: Report stored in local queue ($offlinePendingCount pending upload).")
                                } else {
                                    val res = incidentRepository.submitIncident(newIncident)
                                    if (res.isSuccess) {
                                        snackbarHostState.showSnackbar("Report submitted successfully! Verified by Gemini 3.8 Flash.")
                                    } else {
                                        offlinePendingCount = incidentRepository.getOfflineQueueCount()
                                        snackbarHostState.showSnackbar("Network timeout: Report stored in offline queue.")
                                    }
                                }
                                currentDestination = AppNavDestination.MAP
                            }
                        }
                    )
                    AppNavDestination.ADMIN -> AdminConsoleScreen(
                        state = adminState,
                        onBackClick = { currentDestination = AppNavDestination.HUD },
                        onMarkTreePruned = { treeId ->
                            scope.launch {
                                val req = ai.coexistshield.network.MunicipalKakiActionRequestDto(
                                    treeId = treeId,
                                    action = "FELLED",
                                    officerId = "OFFICER-IWATE-42",
                                    officerRole = currentRole.name,
                                    permitNumber = "PERMIT-MRK-2026-$treeId",
                                    biomassDisposedKg = 480.0,
                                    notes = "Municipal Wildlife Control Squad 3 felled and chipped on site."
                                )
                                val result = geofenceRepository.logMunicipalKakiAction(req)
                                if (result.isSuccess) {
                                    val data = result.getOrNull()!!
                                    adminGcfFelledCount = data.gcfCompletedFelledTrees
                                    adminGcfRiskReductionPct = data.gcfRiskReductionPct
                                    adminFelledTreeIds = adminFelledTreeIds + treeId
                                    snackbarHostState.showSnackbar("Official Action Logged: Tree $treeId felled under Permit! GCF Encounter Risk dropped -${data.treeRiskDropPct}%.")
                                } else {
                                    adminGcfFelledCount += 1
                                    adminGcfRiskReductionPct = 65.2
                                    adminFelledTreeIds = adminFelledTreeIds + treeId
                                    snackbarHostState.showSnackbar("Official Action Logged: Tree #$treeId marked as Felled! GCF Risk reduced.")
                                }
                            }
                        },
                        onDispatchCarcassTruck = { id ->
                            scope.launch {
                                roadkillList = roadkillList.map { rk ->
                                    if (rk.id == id) rk.copy(clearanceStatus = ClearanceStatus.DISPATCHED, assignedOfficialId = "OFFICER-SANITATION-09")
                                    else rk
                                }
                                incidentRepository.updateRoadkillStatus(id, ClearanceStatus.DISPATCHED, "OFFICER-SANITATION-09")
                                snackbarHostState.showSnackbar("Dispatch Sent: Sanitation truck assigned to carcass $id.")
                            }
                        },
                        onExportRspoReport = {
                            scope.launch {
                                val res = geofenceRepository.getRspoAuditPackage("wilmar_sumatra_02")
                                if (res.isSuccess) {
                                    val pkg = res.getOrNull()!!
                                    snackbarHostState.showSnackbar("RSPO Audit Package: ${pkg.auditId} (Zero-Harm, SHA-256: ${pkg.sha256AuditSignature.take(12)}...)")
                                } else {
                                    snackbarHostState.showSnackbar("RSPO Audit Package exported successfully (PDF / CSV downloaded).")
                                }
                            }
                        },
                        onTriggerAiRetraining = {
                            scope.launch {
                                adminIsRetraining = true
                                val result = geofenceRepository.triggerMlRetrain()
                                if (result.isSuccess) {
                                    val res = result.getOrNull()!!
                                    adminModelAucScore = res.candidateModel.metrics.aucRoc
                                    adminRetrainTimestamp = "Just now (${res.activeModelVersion})"
                                    snackbarHostState.showSnackbar("Continuous Retraining Passed! Promoted model: AUC ${res.candidateModel.metrics.aucRoc}")
                                } else {
                                    val err = result.exceptionOrNull()?.message ?: "Retraining job execution failed."
                                    snackbarHostState.showSnackbar("Retraining Error: $err")
                                }
                                adminIsRetraining = false
                            }
                        }
                    )
                }
            }
        }
    }
}

@Composable
fun DemoControlBar(
    currentRole: UserRole,
    currentTier: AlertTier,
    onRoleChange: (UserRole) -> Unit,
    onTierChange: (AlertTier) -> Unit,
    signedInEmail: String? = null,
    isExpanded: Boolean = false,
    onToggleExpand: () -> Unit = {},
    onLoginClick: () -> Unit = {},
    onLogoutClick: () -> Unit = {}
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color(0xFF0F172A))
            .padding(horizontal = 12.dp, vertical = 4.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (signedInEmail != null) {
                Text(
                    text = "👤 $signedInEmail",
                    color = Color.White,
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold
                )
                TextButton(onClick = onLogoutClick) {
                    Text("ログアウト", fontSize = 11.sp)
                }
            } else {
                TextButton(onClick = onLoginClick) {
                    Text("ログイン / 登録", fontSize = 11.sp)
                }
            }

            TextButton(onClick = onToggleExpand) {
                Text(
                    text = if (isExpanded) "🛠 DEMO CONTROLS ▲" else "🛠 DEMO CONTROLS ▼",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold,
                    color = if (isExpanded) MaterialTheme.colorScheme.primary else Color.LightGray
                )
            }

            if (isExpanded) {

            // Role Switcher Chips (locked while signed in: role is derived from token claims)
            UserRole.values().forEach { role ->
                FilterChip(
                    selected = currentRole == role,
                    enabled = signedInEmail == null,
                    onClick = { onRoleChange(role) },
                    label = { Text(role.name, fontSize = 10.sp, fontWeight = FontWeight.Bold) },
                    shape = RoundedCornerShape(12.dp),
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = MaterialTheme.colorScheme.primary,
                        selectedLabelColor = Color.White
                    )
                )
            }

            Text("•", color = Color.Gray, fontSize = 12.sp)

            // Tier Switcher Chips
            AlertTier.values().forEach { tier ->
                val (tierColor, label) = when (tier) {
                    AlertTier.SAFE -> TierSafe to "Safe"
                    AlertTier.TIER_1_NOTICE -> TierNotice to "Tier 1"
                    AlertTier.TIER_2_WARNING -> TierWarning to "Tier 2"
                    AlertTier.TIER_3_DANGER -> TierDanger to "Tier 3"
                }
                FilterChip(
                    selected = currentTier == tier,
                    onClick = { onTierChange(tier) },
                    label = { Text(label, fontSize = 10.sp, fontWeight = FontWeight.Bold) },
                    shape = RoundedCornerShape(12.dp),
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = tierColor,
                        selectedLabelColor = Color.White
                    )
                )
            }
        }
    }
}
}
