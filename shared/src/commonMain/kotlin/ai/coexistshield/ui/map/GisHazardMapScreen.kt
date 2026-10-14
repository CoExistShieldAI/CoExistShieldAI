package ai.coexistshield.ui.map

import ai.coexistshield.model.AlertTier
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
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
    val canMarkResolved: Boolean = false
)

data class GisMapUiState(
    val userLocation: GeoLocation? = GeoLocation(39.702, 141.152),
    val activeGeofences: List<GeofenceZone> = emptyList(),
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
    onOfficialActionClick: (String) -> Unit = {}
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

            // 2. Floating Layer Filter Bar (Top overlay)
            LayerFilterOverlay(
                activeLayers = state.activeLayers,
                onToggleLayer = onToggleLayer,
                modifier = Modifier
                    .align(Alignment.TopCenter)
                    .padding(horizontal = 12.dp, vertical = 8.dp)
            )

            // 3. Entity Details Bottom Sheet / Card (Bottom overlay when pin is selected)
            state.selectedEntity?.let { entity ->
                EntityDetailCard(
                    entity = entity,
                    userRole = state.userRole,
                    onDismiss = { onSelectEntity(null) },
                    onActionClick = { onOfficialActionClick(entity.id) },
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(16.dp)
                )
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
    // Interactive GIS canvas container with mock sample pins representing real coordinates
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
            verticalArrangement = Arrangement.spacedBy(12.dp)
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

            Spacer(modifier = Modifier.height(16.dp))

            // Sample Interactive Pin 1: Abandoned Persimmon Tree (Japan)
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
                                canMarkResolved = true
                            )
                        )
                    }
                )
            }

            // Sample Interactive Pin 2: Roadkill Carcass
            if (state.activeLayers.contains(MapLayerType.ROADKILL)) {
                PinButton(
                    title = "Roadkill Carcass #rk_4412 (Wild Boar)",
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
                                canMarkResolved = true
                            )
                        )
                    }
                )
            }

            // Sample Interactive Pins 3: Sumatra Tiger Depredation & Night Corral (Indonesia)
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
                                canMarkResolved = true
                            )
                        )
                    }
                )
                PinButton(
                    title = "Night Corral #ls_sumatra_02 (Water Buffalo)",
                    color = TierNotice,
                    onClick = {
                        onPinClick(
                            SelectedEntityDetails(
                                id = "ls_sumatra_02",
                                title = "Sungai Rawa Water Buffalo Corral",
                                subtitle = "850m to Forest Edge • Predator-Proof Night Enclosure",
                                speciesOrType = "Bubalus bubalis (Protected Corral)",
                                threatTier = AlertTier.TIER_1_NOTICE,
                                geminiBadgeText = "Corral Factor: 65% Risk Reduction Active",
                                canMarkResolved = false
                            )
                        )
                    }
                )
            }

            // Sample Interactive Pins 4 & 5: Thailand Asian Elephant Crop Attractant Plots (HEC)
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
                                canMarkResolved = true
                            )
                        )
                    }
                )

                PinButton(
                    title = "Jackfruit Odor Corridor #th_plot_kb_02 (Ripening)",
                    color = TierWarning,
                    onClick = {
                        onPinClick(
                            SelectedEntityDetails(
                                id = "th_plot_kb_02",
                                title = "Khao Daeng Jackfruit Grove (8.2 ha)",
                                subtitle = "Kui Buri Buffer • 1100m to Forest Edge",
                                speciesOrType = "Artocarpus heterophyllus (High Volatile Esters)",
                                threatTier = AlertTier.TIER_2_WARNING,
                                geminiBadgeText = "Gemini Vision: Aromatic Ester Emission 0.85 • Dynamic Buffer Expanded",
                                canMarkResolved = true
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
            .padding(horizontal = 16.dp, vertical = 10.dp)
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Box(modifier = Modifier.size(10.dp).background(color, CircleShape))
            Spacer(modifier = Modifier.width(8.dp))
            Text(title, color = Color.White, fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
        }
    }
}

@Composable
fun EntityDetailCard(
    entity: SelectedEntityDetails,
    userRole: UserRole,
    onDismiss: () -> Unit,
    onActionClick: () -> Unit,
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

                if ((userRole == UserRole.OFFICIAL || userRole == UserRole.ADMIN) && entity.canMarkResolved) {
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
