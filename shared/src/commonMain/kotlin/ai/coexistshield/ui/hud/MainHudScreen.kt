package ai.coexistshield.ui.hud

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.network.DispatchRecommendationDto
import ai.coexistshield.ui.map.MapRegion
import ai.coexistshield.ui.theme.Slate100
import ai.coexistshield.ui.theme.Slate700
import ai.coexistshield.ui.theme.Slate800
import ai.coexistshield.ui.theme.Slate900
import ai.coexistshield.ui.theme.TierDanger
import ai.coexistshield.ui.theme.TierNotice
import ai.coexistshield.ui.theme.TierSafe
import ai.coexistshield.ui.theme.TierWarning
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.Canvas
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
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.Icon
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
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlin.math.PI
import kotlin.math.cos
import kotlin.math.sin

/**
 * UI State for the Main HUD Screen.
 */
data class MainHudUiState(
    val currentTier: AlertTier = AlertTier.SAFE,
    val targetSpecies: List<WildlifeSpecies> = emptyList(),
    val closestDistanceMeters: Double? = null,
    val causalExplanation: String = "Clear Zone — Continuous causal sensors indicate safe passage within 5km radius.",
    val isOfflineMode: Boolean = false,
    val isSirenPlaying: Boolean = false,
    val isHapticActive: Boolean = false,
    val suggestedDetour: String? = null,
    val activeRole: UserRole = UserRole.CITIZEN,
    val dispatchRecommendations: List<DispatchRecommendationDto> = emptyList(),
    val hazardLocation: GeoLocation? = null,
    val hazardEntityId: String? = null,
    val hazardRegion: MapRegion? = null,
    val hazardLabel: String? = null
)

/**
 * Filter sections for collapsible detail intelligence.
 */
enum class HudIntelSection(val label: String, val icon: String) {
    AI_CAUSAL("AI Analysis", "🤖"),
    DETOUR("Detour Route", "🔀"),
    SAFETY_PROTOCOL("Safety Rules", "🛡️"),
    MUNICIPAL_TASKS("Tasks", "🎯"),
    DETERRENT_SIREN("Siren & Haptic", "🚨")
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainHudScreen(
    state: MainHudUiState,
    onOpenMapClick: () -> Unit = {},
    onOpenHazardMapClick: (GeoLocation, String?, MapRegion?) -> Unit = { _, _, _ -> },
    onQuickReportClick: () -> Unit = {},
    onOfficialDispatchClick: () -> Unit = {},
    onEmergencySosClick: () -> Unit = {},
    onToggleSirenClick: () -> Unit = {},
    onToggleOfflineClick: () -> Unit = {}
) {
    // Collapsible intel state: By default, keep the HUD clean & minimal.
    var showAllDetails by remember { mutableStateOf(false) }
    var activeIntelFilter by remember { mutableStateOf<HudIntelSection?>(null) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = "CoExist Shield AI",
                            fontWeight = FontWeight.Bold,
                            fontSize = 18.sp
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        RoleBadge(role = state.activeRole)
                    }
                },
                actions = {
                    ConnectionStatusIndicator(
                        isOffline = state.isOfflineMode,
                        onToggleOffline = onToggleOfflineClick
                    )
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(horizontal = 16.dp, vertical = 12.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // 1. Primary Graphical Status Display (Graphical Threat Orb & Proximity Gauge)
            GraphicalStatusDisplay(
                tier = state.currentTier,
                species = state.targetSpecies,
                distanceMeters = state.closestDistanceMeters
            )

            // 2. High-Priority Alert & Direct Map Link Card
            StatusSummaryAndMapLinkCard(
                state = state,
                onOpenHazardMapClick = onOpenHazardMapClick,
                onOpenMapClick = onOpenMapClick,
                onEmergencySosClick = onEmergencySosClick
            )

            // 3. Collapsible Intel Control Bar (Simple by default, expandable on demand)
            IntelExpandControlBar(
                state = state,
                isExpanded = showAllDetails,
                activeFilter = activeIntelFilter,
                onToggleExpandAll = {
                    showAllDetails = !showAllDetails
                    if (!showAllDetails) activeIntelFilter = null
                },
                onSelectFilter = { filter ->
                    if (activeIntelFilter == filter) {
                        activeIntelFilter = null
                    } else {
                        activeIntelFilter = filter
                        showAllDetails = true
                    }
                }
            )

            // 4. On-Demand Expandable Intelligence Content
            AnimatedVisibility(
                visible = showAllDetails || activeIntelFilter != null,
                enter = fadeIn() + expandVertically(),
                exit = fadeOut() + shrinkVertically()
            ) {
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    // Filter: Siren & Haptics (Tier 3 or manually triggered)
                    if ((activeIntelFilter == null || activeIntelFilter == HudIntelSection.DETERRENT_SIREN) &&
                        (state.currentTier == AlertTier.TIER_3_DANGER || state.isSirenPlaying)
                    ) {
                        Tier3SirenHapticsCard(
                            isSirenPlaying = state.isSirenPlaying,
                            isHapticActive = state.isHapticActive,
                            onToggleSirenClick = onToggleSirenClick
                        )
                    }

                    // Filter: Detour Suggestion
                    if ((activeIntelFilter == null || activeIntelFilter == HudIntelSection.DETOUR) &&
                        state.suggestedDetour != null
                    ) {
                        DetourSuggestionCard(
                            detour = state.suggestedDetour,
                            onViewOnMap = {
                                if (state.hazardLocation != null) {
                                    onOpenHazardMapClick(state.hazardLocation, state.hazardEntityId, state.hazardRegion)
                                } else {
                                    onOpenMapClick()
                                }
                            }
                        )
                    }

                    // Filter: AI Causal Explanation
                    if (activeIntelFilter == null || activeIntelFilter == HudIntelSection.AI_CAUSAL) {
                        CausalExplanationCard(
                            explanation = state.causalExplanation,
                            tier = state.currentTier
                        )
                    }

                    // Filter: Species-Specific Safety Guidance Protocol
                    if ((activeIntelFilter == null || activeIntelFilter == HudIntelSection.SAFETY_PROTOCOL) &&
                        state.targetSpecies.isNotEmpty()
                    ) {
                        SpeciesSafetyInstructionsCard(species = state.targetSpecies)
                    }

                    // Filter: Municipal Action Tasks (Dispatch recommendations for official/admin)
                    if ((activeIntelFilter == null || activeIntelFilter == HudIntelSection.MUNICIPAL_TASKS) &&
                        state.dispatchRecommendations.isNotEmpty() &&
                        (state.activeRole == UserRole.OFFICIAL || state.activeRole == UserRole.ADMIN)
                    ) {
                        ActionableDispatchRecommendationsCard(
                            recommendations = state.dispatchRecommendations
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // 5. Primary Actions Footer
            ActionFooter(
                role = state.activeRole,
                tier = state.currentTier,
                onOpenMapClick = onOpenMapClick,
                onQuickReportClick = onQuickReportClick,
                onOfficialDispatchClick = onOfficialDispatchClick
            )
        }
    }
}

/**
 * Graphical Status Display:
 * A stylish, circular radial gauge & threat orb that makes safety vs danger immediately apparent at a single glance.
 */
@Composable
fun GraphicalStatusDisplay(
    tier: AlertTier,
    species: List<WildlifeSpecies>,
    distanceMeters: Double?
) {
    val (tierColor, tierLabel, tierIcon, tierDescription) = when (tier) {
        AlertTier.SAFE -> Quadruple(
            TierSafe,
            "SAFE ZONE",
            "🛡️",
            "All clear • No active predators within 5km"
        )
        AlertTier.TIER_1_NOTICE -> Quadruple(
            TierNotice,
            "TIER 1 NOTICE",
            "⚠️",
            "Approaching moderate wildlife corridor"
        )
        AlertTier.TIER_2_WARNING -> Quadruple(
            TierWarning,
            "TIER 2 WARNING",
            "⚡",
            "Active scent plume & encounter hazard"
        )
        AlertTier.TIER_3_DANGER -> Quadruple(
            TierDanger,
            "IMMEDIATE DANGER",
            "🚨",
            "Apex predator / fresh kill-site nearby"
        )
    }

    val animatedColor by animateColorAsState(targetValue = tierColor, animationSpec = tween(500))

    // Pulse animation for warning and danger states
    val infiniteTransition = rememberInfiniteTransition()
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 0.95f,
        targetValue = 1.05f,
        animationSpec = infiniteRepeatable(
            animation = tween(if (tier == AlertTier.TIER_3_DANGER) 500 else 1000),
            repeatMode = RepeatMode.Reverse
        )
    )
    val ringRotation by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 360f,
        animationSpec = infiniteRepeatable(
            animation = tween(12000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        )
    )

    // Calculate proximity gauge percentage (0.0 = far, 1.0 = immediate proximity)
    val proximityRatio = if (distanceMeters != null) {
        ((2000.0 - distanceMeters.coerceIn(50.0, 2000.0)) / 1950.0).toFloat()
    } else if (tier == AlertTier.SAFE) {
        0.0f
    } else {
        0.5f
    }

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.padding(vertical = 8.dp)
    ) {
        Box(
            modifier = Modifier
                .size(220.dp)
                .clip(CircleShape),
            contentAlignment = Alignment.Center
        ) {
            // High-Tech Canvas HUD Gauge Background & Animated Rings
            Canvas(modifier = Modifier.fillMaxSize()) {
                val strokeWidthPx = 10.dp.toPx()
                val radius = (size.minDimension - strokeWidthPx) / 2f
                val centerOffset = Offset(size.width / 2f, size.height / 2f)

                // 1. Ambient Background Glow
                drawCircle(
                    brush = Brush.radialGradient(
                        colors = listOf(
                            animatedColor.copy(alpha = if (tier == AlertTier.TIER_3_DANGER) 0.28f else 0.14f),
                            Color.Transparent
                        ),
                        center = centerOffset,
                        radius = size.width / 1.5f
                    ),
                    radius = radius
                )

                // 2. Base subtle track ring
                drawCircle(
                    color = animatedColor.copy(alpha = 0.18f),
                    radius = radius,
                    style = Stroke(width = strokeWidthPx)
                )

                // 3. Dynamic Proximity Arc (fills based on threat closeness)
                val arcSweep = if (tier == AlertTier.SAFE) {
                    360f // Safe ring is continuous and calm
                } else {
                    (proximityRatio * 320f).coerceIn(45f, 320f)
                }
                val startAngle = -90f

                drawArc(
                    brush = Brush.sweepGradient(
                        colors = listOf(
                            animatedColor.copy(alpha = 0.4f),
                            animatedColor,
                            animatedColor.copy(alpha = 0.9f)
                        ),
                        center = centerOffset
                    ),
                    startAngle = startAngle,
                    sweepAngle = arcSweep,
                    useCenter = false,
                    style = Stroke(width = strokeWidthPx, cap = StrokeCap.Round)
                )

                // 4. Tick markers around circumference for tactical look
                val tickCount = 24
                for (i in 0 until tickCount) {
                    val angleDeg = (i * (360f / tickCount) + ringRotation * 0.1f)
                    val angleRad = angleDeg * (PI / 180f).toFloat()
                    val innerR = radius - 14.dp.toPx()
                    val outerR = radius - 8.dp.toPx()

                    val startX = centerOffset.x + innerR * cos(angleRad)
                    val startY = centerOffset.y + innerR * sin(angleRad)
                    val endX = centerOffset.x + outerR * cos(angleRad)
                    val endY = centerOffset.y + outerR * sin(angleRad)

                    val isMajor = i % 6 == 0
                    drawLine(
                        color = if (isMajor) animatedColor.copy(alpha = 0.7f) else animatedColor.copy(alpha = 0.2f),
                        start = Offset(startX, startY),
                        end = Offset(endX, endY),
                        strokeWidth = if (isMajor) 2.5.dp.toPx() else 1.dp.toPx()
                    )
                }
            }

            // Central Information Orb
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
                modifier = Modifier.padding(16.dp)
            ) {
                // Tier Icon with pulse
                Text(
                    text = tierIcon,
                    fontSize = if (tier == AlertTier.TIER_3_DANGER) 36.sp else 32.sp,
                    modifier = Modifier.padding(bottom = 2.dp)
                )

                // Tier Status Header
                Text(
                    text = tierLabel,
                    fontWeight = FontWeight.Black,
                    fontSize = 17.sp,
                    color = animatedColor,
                    textAlign = TextAlign.Center
                )

                // Distance or Safe state
                if (distanceMeters != null) {
                    Text(
                        text = "${distanceMeters.toInt()}m",
                        fontWeight = FontWeight.ExtraBold,
                        fontSize = 24.sp,
                        color = MaterialTheme.colorScheme.onSurface
                    )
                    Text(
                        text = "ESTIMATED PROXIMITY",
                        fontWeight = FontWeight.Bold,
                        fontSize = 9.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                } else {
                    Text(
                        text = "ALL SECURE",
                        fontWeight = FontWeight.ExtraBold,
                        fontSize = 18.sp,
                        color = MaterialTheme.colorScheme.onSurface
                    )
                    Text(
                        text = "CONTINUOUS SENSING",
                        fontWeight = FontWeight.Bold,
                        fontSize = 9.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }

                // Fauna Name if available
                if (species.isNotEmpty()) {
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = species.joinToString(" • ") { it.name.replace("_", " ") },
                        fontWeight = FontWeight.SemiBold,
                        fontSize = 11.sp,
                        color = animatedColor,
                        maxLines = 1
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(6.dp))

        // Subtitle status description
        Text(
            text = tierDescription,
            fontSize = 13.sp,
            fontWeight = FontWeight.Medium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f),
            textAlign = TextAlign.Center,
            modifier = Modifier.padding(horizontal = 24.dp)
        )
    }
}

/**
 * Status Summary & Direct Map Navigation Link:
 * Provides immediate feedback on why the alert is active and a 1-click button to view the exact hazard on the GIS map.
 */
@Composable
fun StatusSummaryAndMapLinkCard(
    state: MainHudUiState,
    onOpenHazardMapClick: (GeoLocation, String?, MapRegion?) -> Unit,
    onOpenMapClick: () -> Unit,
    onEmergencySosClick: () -> Unit
) {
    val tierColor = when (state.currentTier) {
        AlertTier.SAFE -> TierSafe
        AlertTier.TIER_1_NOTICE -> TierNotice
        AlertTier.TIER_2_WARNING -> TierWarning
        AlertTier.TIER_3_DANGER -> TierDanger
    }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, tierColor.copy(alpha = 0.4f), RoundedCornerShape(16.dp)),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            // Direct Link to GIS Map for this hazard
            if (state.currentTier != AlertTier.SAFE) {
                Button(
                    onClick = {
                        if (state.hazardLocation != null) {
                            onOpenHazardMapClick(state.hazardLocation, state.hazardEntityId, state.hazardRegion)
                        } else {
                            onOpenMapClick()
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(48.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = tierColor
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.Center
                    ) {
                        Text("📍", fontSize = 16.sp)
                        Spacer(modifier = Modifier.width(8.dp))
                        val distanceText = state.closestDistanceMeters?.let { " (${it.toInt()}m ahead)" } ?: ""
                        Text(
                            text = "View Hazard Location on Map$distanceText",
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                }
            }

            // Emergency SOS for Tier 3 Immediate Danger
            if (state.currentTier == AlertTier.TIER_3_DANGER) {
                OutlinedButton(
                    onClick = onEmergencySosClick,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp),
                    colors = ButtonDefaults.outlinedButtonColors(
                        contentColor = TierDanger
                    ),
                    border = androidx.compose.foundation.BorderStroke(
                        1.dp,
                        Brush.horizontalGradient(listOf(TierDanger, Color(0xFFB91C1C)))
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text(
                        text = "🚨 BROADCAST EMERGENCY SOS DISPATCH",
                        fontWeight = FontWeight.Black,
                        fontSize = 12.sp,
                        color = TierDanger
                    )
                }
            }

            // Brief summary tag
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = if (state.currentTier == AlertTier.SAFE) "Autonomous Sensor Perimeter: Nominal" else "Active Threat Perimeter Detected",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                )
                if (state.hazardRegion != null) {
                    Text(
                        text = state.hazardRegion.label,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
        }
    }
}

/**
 * Collapsible Intel Control Bar:
 * Allows user to view detailed intelligence on demand using convenient category chips or an expand-all toggle.
 */
@Composable
fun IntelExpandControlBar(
    state: MainHudUiState,
    isExpanded: Boolean,
    activeFilter: HudIntelSection?,
    onToggleExpandAll: () -> Unit,
    onSelectFilter: (HudIntelSection) -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = "Additional Intelligence",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Spacer(modifier = Modifier.width(6.dp))
                Box(
                    modifier = Modifier
                        .background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(10.dp))
                        .padding(horizontal = 6.dp, vertical = 2.dp)
                ) {
                    Text(
                        text = if (isExpanded) "Active" else "Collapsed",
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Medium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            FilledTonalButton(
                onClick = onToggleExpandAll,
                shape = RoundedCornerShape(8.dp),
                colors = ButtonDefaults.filledTonalButtonColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Text(
                    text = if (isExpanded) "Hide Details ▲" else "View Details ▼",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold
                )
            }
        }

        // Quick Category Filter Chips
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            HudIntelSection.values().forEach { section ->
                // Skip tasks for non-officials
                if (section == HudIntelSection.MUNICIPAL_TASKS &&
                    state.activeRole != UserRole.OFFICIAL && state.activeRole != UserRole.ADMIN
                ) {
                    return@forEach
                }

                // Skip detour if no detour available
                if (section == HudIntelSection.DETOUR && state.suggestedDetour == null) {
                    return@forEach
                }

                // Skip siren if not Tier 3 or playing
                if (section == HudIntelSection.DETERRENT_SIREN &&
                    state.currentTier != AlertTier.TIER_3_DANGER && !state.isSirenPlaying
                ) {
                    return@forEach
                }

                val isSelected = activeFilter == section
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(12.dp))
                        .background(
                            if (isSelected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
                        )
                        .border(
                            1.dp,
                            if (isSelected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.outline.copy(alpha = 0.2f),
                            RoundedCornerShape(12.dp)
                        )
                        .clickable { onSelectFilter(section) }
                        .padding(horizontal = 10.dp, vertical = 6.dp)
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(section.icon, fontSize = 12.sp)
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = section.label,
                            fontSize = 11.sp,
                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                            color = if (isSelected) Color.White else MaterialTheme.colorScheme.onSurface
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun RoleBadge(role: UserRole) {
    val (label, bg, fg) = when (role) {
        UserRole.CITIZEN -> Triple("CITIZEN", Color(0xFFE2E8F0), Color(0xFF334155))
        UserRole.OFFICIAL -> Triple("OFFICIAL", Color(0xFFDBEAFE), Color(0xFF1E40AF))
        UserRole.ADMIN -> Triple("ADMIN", Color(0xFFFCE7F3), Color(0xFF9D174D))
    }
    Box(
        modifier = Modifier
            .background(bg, RoundedCornerShape(12.dp))
            .padding(horizontal = 8.dp, vertical = 2.dp)
    ) {
        Text(text = label, color = fg, fontSize = 11.sp, fontWeight = FontWeight.SemiBold)
    }
}

@Composable
fun ConnectionStatusIndicator(
    isOffline: Boolean,
    onToggleOffline: () -> Unit = {}
) {
    Row(
        modifier = Modifier
            .clip(RoundedCornerShape(8.dp))
            .clickable(onClick = onToggleOffline)
            .padding(horizontal = 8.dp, vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(8.dp)
                .background(
                    if (isOffline) Color.Gray else TierSafe,
                    CircleShape
                )
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = if (isOffline) "Offline (Cached)" else "Live",
            fontSize = 12.sp,
            fontWeight = FontWeight.Medium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
        )
    }
}

@Composable
fun DetourSuggestionCard(
    detour: String,
    onViewOnMap: () -> Unit = {}
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.7f)),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("🔀", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(10.dp))
                Column {
                    Text(
                        text = "Recommended Safety Detour Route",
                        fontWeight = FontWeight.Bold,
                        fontSize = 13.sp,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = detour,
                        fontSize = 12.sp,
                        lineHeight = 16.sp,
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }
            }
            OutlinedButton(
                onClick = onViewOnMap,
                modifier = Modifier.fillMaxWidth().height(36.dp),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text("🗺️ View Detour Corridor on Map", fontSize = 11.sp, fontWeight = FontWeight.SemiBold)
            }
        }
    }
}

@Composable
fun CausalExplanationCard(
    explanation: String,
    tier: AlertTier
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.6f)
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = "AI Causal Reasoning",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "• Powered by Gemini 3.8 Flash",
                    fontSize = 11.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
            }
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = explanation,
                fontSize = 13.sp,
                lineHeight = 19.sp,
                color = MaterialTheme.colorScheme.onSurface
            )
        }
    }
}

@Composable
fun SpeciesSafetyInstructionsCard(species: List<WildlifeSpecies>) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("🛡️", fontSize = 16.sp)
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "Species Safety & Deterrent Protocol",
                    fontWeight = FontWeight.Bold,
                    fontSize = 13.sp,
                    color = MaterialTheme.colorScheme.onSurface
                )
            }

            species.forEach { sp ->
                val (name, protocol) = when (sp) {
                    WildlifeSpecies.BEAR -> "Asian Black Bear (Ursus thibetanus)" to
                        "1. Make presence known with voice / bear bell.\n2. Do NOT run or climb trees; stand ground.\n3. Back away slowly facing animal.\n4. Arm capsaicin spray."
                    WildlifeSpecies.TIGER -> "Sumatran Tiger (Panthera tigris sumatrae)" to
                        "1. Maintain upright posture; never crouch.\n2. Do NOT turn your back or run.\n3. Keep continuous eye contact.\n4. Retreat slowly to vehicle or night corral."
                    WildlifeSpecies.ELEPHANT -> "Asian Elephant (Elephas maximus)" to
                        "1. Dim headlights and do NOT sound vehicle horns.\n2. Maintain minimum 100m standoff buffer.\n3. Do NOT block migration trails to water or crop buffers."
                    else -> sp.name to "Maintain distance and avoid confrontation."
                }

                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f), RoundedCornerShape(8.dp))
                        .padding(10.dp)
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Text(text = name, fontWeight = FontWeight.Bold, fontSize = 12.sp, color = MaterialTheme.colorScheme.primary)
                        Text(text = protocol, fontSize = 11.sp, lineHeight = 16.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }
    }
}

@Composable
fun Tier3SirenHapticsCard(
    isSirenPlaying: Boolean,
    isHapticActive: Boolean,
    onToggleSirenClick: () -> Unit
) {
    val infiniteTransition = rememberInfiniteTransition()
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.6f,
        targetValue = 1.0f,
        animationSpec = infiniteRepeatable(
            animation = tween(400),
            repeatMode = RepeatMode.Reverse
        )
    )

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(2.dp, TierDanger.copy(alpha = pulseAlpha), RoundedCornerShape(14.dp)),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = TierDanger.copy(alpha = 0.15f))
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("🚨", fontSize = 20.sp)
                    Spacer(modifier = Modifier.width(8.dp))
                    Column {
                        Text(
                            text = "ACOUSTIC & HAPTIC ALERT",
                            color = TierDanger,
                            fontWeight = FontWeight.Black,
                            fontSize = 14.sp
                        )
                        Text(
                            text = if (isSirenPlaying) "High-Frequency Deterrent Siren Emitting (95dB)" else "Acoustic Deterrent Armed",
                            fontSize = 11.sp,
                            color = MaterialTheme.colorScheme.onSurface
                        )
                    }
                }

                Button(
                    onClick = onToggleSirenClick,
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isSirenPlaying) Color.DarkGray else TierDanger
                    ),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = if (isSirenPlaying) "MUTE SIREN" else "TRIGGER SIREN",
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                }
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                    .padding(horizontal = 10.dp, vertical = 6.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "📳 Haptic Pulse Feedback",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = if (isHapticActive) "VIBRATING (Rapid Pulse)" else "ACTIVE",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Bold,
                    color = TierDanger
                )
            }
        }
    }
}

@Composable
fun ActionableDispatchRecommendationsCard(
    recommendations: List<DispatchRecommendationDto>
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("🎯", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "Prioritized Municipal Action Tasks",
                        fontWeight = FontWeight.Bold,
                        fontSize = 13.sp,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                Box(
                    modifier = Modifier
                        .background(TierDanger.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 2.dp)
                ) {
                    Text("AGENT DISPATCH", color = TierDanger, fontSize = 10.sp, fontWeight = FontWeight.Bold)
                }
            }

            recommendations.forEachIndexed { idx, rec ->
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                        .padding(10.dp)
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = "${idx + 1}. [${rec.actionType}] ${rec.location}",
                                fontWeight = FontWeight.Bold,
                                fontSize = 12.sp,
                                color = MaterialTheme.colorScheme.onSurface
                            )
                            Box(
                                modifier = Modifier
                                    .background(
                                        if (rec.priority == "IMMEDIATE") TierDanger else TierWarning,
                                        RoundedCornerShape(4.dp)
                                    )
                                    .padding(horizontal = 6.dp, vertical = 2.dp)
                            ) {
                                Text(
                                    text = rec.priority,
                                    fontSize = 9.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color.White
                                )
                            }
                        }
                        Text(
                            text = rec.justification,
                            fontSize = 11.sp,
                            lineHeight = 15.sp,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun ActionFooter(
    role: UserRole,
    tier: AlertTier,
    onOpenMapClick: () -> Unit,
    onQuickReportClick: () -> Unit,
    onOfficialDispatchClick: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        Button(
            onClick = onOpenMapClick,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text(text = "Open Interactive GIS Map", fontSize = 16.sp, fontWeight = FontWeight.Bold)
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            OutlinedButton(
                onClick = onQuickReportClick,
                modifier = Modifier
                    .weight(1f)
                    .height(48.dp),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Camera Report", fontSize = 14.sp)
            }

            if (role == UserRole.OFFICIAL || role == UserRole.ADMIN) {
                FilledTonalButton(
                    onClick = onOfficialDispatchClick,
                    modifier = Modifier
                        .weight(1f)
                        .height(48.dp),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Official Console", fontSize = 14.sp)
                }
            }
        }
    }
}

data class Quadruple<A, B, C, D>(val first: A, val second: B, val third: C, val fourth: D)
