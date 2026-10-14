package ai.coexistshield.ui.map

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.ClearanceStatus
import ai.coexistshield.model.DecayState
import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.KakiTree
import ai.coexistshield.model.RoadkillRecord
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.ui.theme.Slate900
import ai.coexistshield.ui.theme.TierDanger
import ai.coexistshield.ui.theme.TierNotice
import ai.coexistshield.ui.theme.TierSafe
import ai.coexistshield.ui.theme.TierWarning
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
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
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class MapLayerType(val label: String) {
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
    val roadkillStatus: ClearanceStatus? = null
)

data class GisMapUiState(
    val userLocation: GeoLocation? = GeoLocation(39.702, 141.152),
    val activeGeofences: List<GeofenceZone> = emptyList(),
    val roadkills: List<RoadkillRecord> = emptyList(),
    val proximityRoadkillAlert: RoadkillRecord? = null,
    val activeLayers: Set<MapLayerType> = setOf(
        MapLayerType.ROADKILL,
        MapLayerType.PERSIMMON_TREES,
        MapLayerType.ELEPHANT_CROPS,
        MapLayerType.LIVESTOCK_PREDATION
    ),
    val selectedEntity: SelectedEntityDetails? = null,
    val userRole: UserRole = UserRole.CITIZEN
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GisHazardMapScreen(
    state: GisMapUiState,
    onBackClick: () -> Unit = {},
    onToggleLayer: (MapLayerType) -> Unit = {},
    onSelectEntity: (SelectedEntityDetails?) -> Unit = {},
    onOfficialActionClick: (String) -> Unit = {},
    onUpdateRoadkillStatus: (String, ClearanceStatus) -> Unit = { _, _ -> }
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("GIS Hazard & Attractant Map", fontSize = 18.sp, fontWeight = FontWeight.Bold)
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
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
        ) {
            // 1. Google Maps Vector Canvas Container
            MapCanvasPlaceholder(
                state = state,
                onPinClick = onSelectEntity
            )

            // 2. Top Overlays: Layer Filter Bar + Proximity Warning Banner
            Column(
                modifier = Modifier
                    .align(Alignment.TopCenter)
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                LayerFilterOverlay(
                    activeLayers = state.activeLayers,
                    onToggleLayer = onToggleLayer
                )

                // Issue #3: Proximity Warning Banner (< 800m to decomposing carcass)
                val activeCarcassAlert = state.proximityRoadkillAlert ?: state.roadkills.firstOrNull {
                    it.clearanceStatus != ClearanceStatus.CLEARED && it.attractantThreatScore >= 0.70
                }
                if (activeCarcassAlert != null && state.activeLayers.contains(MapLayerType.ROADKILL)) {
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
                        }
                    )
                }
            }

            // 3. Entity Details Bottom Sheet / Card (Bottom overlay when pin is selected)
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
fun RoadkillProximityWarningBanner(
    roadkill: RoadkillRecord,
    onClick: () -> Unit
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
            .border(2.dp, TierDanger.copy(alpha = pulseAlpha), RoundedCornerShape(12.dp))
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = TierDanger.copy(alpha = 0.95f))
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("⚠️", fontSize = 20.sp)
            Spacer(modifier = Modifier.width(10.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "ROADKILL PROXIMITY ALERT (< 800m)",
                    color = Color.White,
                    fontWeight = FontWeight.Black,
                    fontSize = 13.sp
                )
                Text(
                    text = "${roadkill.species.name.replace("_", " ")} (${roadkill.decayState.name} Carcass) • Threat Score: ${roadkill.attractantThreatScore}. High apex predator attractant lure. Avoid stopping.",
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 11.sp,
                    lineHeight = 15.sp
                )
            }
            Spacer(modifier = Modifier.width(6.dp))
            Box(
                modifier = Modifier
                    .background(Color.White, RoundedCornerShape(6.dp))
                    .padding(horizontal = 8.dp, vertical = 4.dp)
            ) {
                Text("INSPECT", fontSize = 10.sp, fontWeight = FontWeight.Bold, color = TierDanger)
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
                label = { Text(layer.label, fontSize = 12.sp) },
                shape = RoundedCornerShape(20.dp),
                colors = FilterChipDefaults.filterChipColors(
                    selectedContainerColor = MaterialTheme.colorScheme.primaryContainer,
                    selectedLabelColor = MaterialTheme.colorScheme.onPrimaryContainer
                )
            )
        }
    }
}

@Composable
fun MapCanvasPlaceholder(
    state: GisMapUiState,
    onPinClick: (SelectedEntityDetails) -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF1E293B))
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.Center)
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            Text(
                text = "Google Maps Vector Engine Active",
                color = Color.White,
                fontWeight = FontWeight.Bold,
                fontSize = 16.sp
            )
            Text(
                text = "Rendering Dynamic GeoJSON Polygons & Attractant Layers",
                color = Color.LightGray,
                fontSize = 12.sp
            )

            Spacer(modifier = Modifier.height(10.dp))

            // Roadkill Layer Pins with Lifecycle Badges (REPORTED / DISPATCHED / CLEARED)
            if (state.activeLayers.contains(MapLayerType.ROADKILL)) {
                if (state.roadkills.isNotEmpty()) {
                    state.roadkills.forEach { rk ->
                        val (chipColor, statusLabel) = when (rk.clearanceStatus) {
                            ClearanceStatus.REPORTED -> TierWarning to "REPORTED (Active Lure)"
                            ClearanceStatus.DISPATCHED -> Color(0xFF38BDF8) to "DISPATCHED (Sanitation En Route)"
                            ClearanceStatus.CLEARED -> TierSafe to "CLEARED (Neutralized)"
                        }

                        PinButton(
                            title = "Roadkill #${rk.id} (${rk.species.name.replace("_", " ")}) • $statusLabel",
                            color = chipColor,
                            onClick = {
                                onPinClick(
                                    SelectedEntityDetails(
                                        id = rk.id,
                                        title = "Roadkill: ${rk.species.name.replace("_", " ")}",
                                        subtitle = "Location: Lat ${rk.location.latitude}, Lon ${rk.location.longitude} • State: ${rk.decayState.name}",
                                        speciesOrType = rk.species.name,
                                        threatTier = if (rk.clearanceStatus == ClearanceStatus.CLEARED) AlertTier.SAFE else AlertTier.TIER_2_WARNING,
                                        geminiBadgeText = "Gemini Forensics: Threat Index ${rk.attractantThreatScore} • Status ${rk.clearanceStatus.name}",
                                        canMarkResolved = true,
                                        isRoadkill = true,
                                        roadkillStatus = rk.clearanceStatus
                                    )
                                )
                            }
                        )
                    }
                } else {
                    // Fallback sample roadkill pins if list is not passed
                    PinButton(
                        title = "Roadkill Carcass #rk_4412 (Wild Boar) • REPORTED",
                        color = TierWarning,
                        onClick = {
                            onPinClick(
                                SelectedEntityDetails(
                                    id = "rk_4412",
                                    title = "Decomposing Wild Boar Carcass",
                                    subtitle = "Route 4 Northbound • Active Attractant Magnet",
                                    speciesOrType = "Sus scrofa",
                                    threatTier = AlertTier.TIER_2_WARNING,
                                    geminiBadgeText = "Gemini Forensics: Bloated Decay • Threat Index 0.88",
                                    canMarkResolved = true,
                                    isRoadkill = true,
                                    roadkillStatus = ClearanceStatus.REPORTED
                                )
                            )
                        }
                    )
                }
            }

            // Kaki Trees Layer (Japan)
            if (state.activeLayers.contains(MapLayerType.PERSIMMON_TREES)) {
                PinButton(
                    title = "Kaki Tree #kt_1120 (Top 5% Hazard)",
                    color = TierDanger,
                    onClick = {
                        onPinClick(
                            SelectedEntityDetails(
                                id = "kt_1120",
                                title = "Abandoned Persimmon Tree",
                                subtitle = "Vacant House • Near Elementary School Route (120m)",
                                speciesOrType = "Diospyros kaki",
                                threatTier = AlertTier.TIER_2_WARNING,
                                geminiBadgeText = "Gemini Vision: Abundant Fruit Load • Overripe Falling",
                                canMarkResolved = true,
                                isRoadkill = false
                            )
                        )
                    }
                )
            }

            // Sumatra Tiger Kill-Site Layer (Indonesia)
            if (state.activeLayers.contains(MapLayerType.LIVESTOCK_PREDATION)) {
                PinButton(
                    title = "Tiger Kill-Site #ls_sumatra_01 (72h Caching)",
                    color = TierDanger,
                    onClick = {
                        onPinClick(
                            SelectedEntityDetails(
                                id = "ls_sumatra_01",
                                title = "Bukit Kesuma Cattle Kill-Site (Sumatra)",
                                subtitle = "320m to Forest Edge • Active 72h Feeding Return Cycle",
                                speciesOrType = "Panthera tigris sumatrae (Sumatran Tiger)",
                                threatTier = AlertTier.TIER_3_DANGER,
                                geminiBadgeText = "Gemini Forensics: 72mm Canine Puncture Match • 3km Village Alert Active",
                                canMarkResolved = true,
                                isRoadkill = false
                            )
                        )
                    }
                )
            }

            // Thailand Asian Elephant Crop Attractant Plots (HEC)
            if (state.activeLayers.contains(MapLayerType.ELEPHANT_CROPS)) {
                PinButton(
                    title = "Elephant Attractant #th_plot_kb_01 (Sugarcane)",
                    color = TierDanger,
                    onClick = {
                        onPinClick(
                            SelectedEntityDetails(
                                id = "th_plot_kb_01",
                                title = "Kui Buri Sugarcane Estate #1 (14.5 ha)",
                                subtitle = "Prachuap Khiri Khan • 650m to National Park Border",
                                speciesOrType = "Saccharum officinarum (Elephant Calorie Magnet)",
                                threatTier = AlertTier.TIER_3_DANGER,
                                geminiBadgeText = "Gemini Vision: High Sucrose Spike • 2.8km Scent Corridor blowing West into Reserve",
                                canMarkResolved = true,
                                isRoadkill = false
                            )
                        )
                    }
                )
            }
        }
    }
}

@Composable
fun PinButton(
    title: String,
    color: Color,
    onClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(8.dp))
            .background(color.copy(alpha = 0.2f))
            .border(1.dp, color, RoundedCornerShape(8.dp))
            .clickable(onClick = onClick)
            .padding(horizontal = 14.dp, vertical = 8.dp)
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(10.dp).background(color, CircleShape))
            Spacer(modifier = Modifier.width(8.dp))
            Text(title, color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
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
            containerColor = MaterialTheme.colorScheme.surface
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
                    fontSize = 16.sp
                )
                Box(
                    modifier = Modifier
                        .clickable(onClick = onDismiss)
                        .padding(4.dp)
                ) {
                    Text("✕", fontSize = 14.sp, color = Color.Gray)
                }
            }

            Text(
                text = entity.subtitle,
                fontSize = 13.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
            )

            // Status chip for roadkill
            if (entity.isRoadkill && entity.roadkillStatus != null) {
                val (chipColor, label) = when (entity.roadkillStatus) {
                    ClearanceStatus.REPORTED -> TierWarning to "STATUS: REPORTED (Pending Sanitation Dispatch)"
                    ClearanceStatus.DISPATCHED -> Color(0xFF0284C7) to "STATUS: DISPATCHED (Truck Assigned)"
                    ClearanceStatus.CLEARED -> TierSafe to "STATUS: CLEARED (Mitigation Complete)"
                }
                Box(
                    modifier = Modifier
                        .background(chipColor.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
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
                        .background(MaterialTheme.colorScheme.primaryContainer, RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = badge,
                        color = MaterialTheme.colorScheme.onPrimaryContainer,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium
                    )
                }
            }

            Spacer(modifier = Modifier.height(6.dp))

            // Action row
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedButton(
                    onClick = onDismiss,
                    modifier = Modifier.weight(1f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text("Close", fontSize = 13.sp)
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
                                Text("Dispatch Sanitation Truck", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                        ClearanceStatus.DISPATCHED -> {
                            Button(
                                onClick = { onRoadkillStatusChange(ClearanceStatus.CLEARED) },
                                modifier = Modifier.weight(1.8f),
                                shape = RoundedCornerShape(8.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = TierSafe)
                            ) {
                                Text("Confirm Carcass Cleared", fontSize = 12.sp, fontWeight = FontWeight.Bold)
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
                        Text("Official: Mark Resolved", fontSize = 13.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}
