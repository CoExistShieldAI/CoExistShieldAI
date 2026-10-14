package ai.coexistshield.ui.hud

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
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

data class MainHudUiState(
    val currentTier: AlertTier = AlertTier.SAFE,
    val targetSpecies: List<WildlifeSpecies> = emptyList(),
    val closestDistanceMeters: Double? = null,
    val causalExplanation: String = "Clear Zone — Continuous causal sensors indicate safe passage within 5km radius.",
    val isOfflineMode: Boolean = false,
    val isSirenPlaying: Boolean = false,
    val isHapticActive: Boolean = false,
    val suggestedDetour: String? = null,
    val activeRole: UserRole = UserRole.CITIZEN
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainHudScreen(
    state: MainHudUiState,
    onOpenMapClick: () -> Unit = {},
    onQuickReportClick: () -> Unit = {},
    onOfficialDispatchClick: () -> Unit = {},
    onEmergencySosClick: () -> Unit = {},
    onToggleSirenClick: () -> Unit = {},
    onToggleOfflineClick: () -> Unit = {}
) {
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
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // 1. Primary Dynamic Alert Banner
            AlertBannerCard(
                tier = state.currentTier,
                species = state.targetSpecies,
                distanceMeters = state.closestDistanceMeters,
                onEmergencySosClick = onEmergencySosClick
            )

            // 2. Spatial Distance Tracking Gauge Card
            if (state.closestDistanceMeters != null) {
                SpatialDistanceTrackingCard(
                    distanceMeters = state.closestDistanceMeters,
                    tier = state.currentTier
                )
            }

            // 3. Tier 3 Acoustic Siren & Haptic Alert Card
            if (state.currentTier == AlertTier.TIER_3_DANGER) {
                Tier3SirenHapticsCard(
                    isSirenPlaying = state.isSirenPlaying,
                    isHapticActive = state.isHapticActive,
                    onToggleSirenClick = onToggleSirenClick
                )
            }

            // 4. Detour Route Suggestion (Tier 2 & 3)
            if (state.suggestedDetour != null) {
                DetourSuggestionCard(detour = state.suggestedDetour)
            }

            // 5. AI Causal Explanation (Explainable AI Grounding via Gemini 3.8 Flash)
            CausalExplanationCard(
                explanation = state.causalExplanation,
                tier = state.currentTier
            )

            // 6. Species-Specific Safety Guidance Protocol
            if (state.targetSpecies.isNotEmpty()) {
                SpeciesSafetyInstructionsCard(species = state.targetSpecies)
            }

            // 7. Active Target Fauna Card
            if (state.targetSpecies.isNotEmpty()) {
                SpeciesThreatCard(
                    species = state.targetSpecies,
                    distanceMeters = state.closestDistanceMeters
                )
            }

            Spacer(modifier = Modifier.weight(1f, fill = false))

            // 8. Primary Actions Footer
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
fun AlertBannerCard(
    tier: AlertTier,
    species: List<WildlifeSpecies>,
    distanceMeters: Double?,
    onEmergencySosClick: () -> Unit
) {
    val (tierColor, title, subtitle) = when (tier) {
        AlertTier.SAFE -> Triple(
            TierSafe,
            "SAFE ZONE",
            "No active predator threats detected nearby."
        )
        AlertTier.TIER_1_NOTICE -> Triple(
            TierNotice,
            "TIER 1 : NOTICE",
            "Heightened vigilance advised. Approaching moderate wildlife corridor."
        )
        AlertTier.TIER_2_WARNING -> Triple(
            TierWarning,
            "TIER 2 : WARNING",
            "High encounter probability. Scent plume / attractant corridor active."
        )
        AlertTier.TIER_3_DANGER -> Triple(
            TierDanger,
            "TIER 3 : IMMEDIATE DANGER",
            "IMMINENT HAZARD. Active apex predator or fresh kill-site nearby."
        )
    }

    val animatedColor by animateColorAsState(targetValue = tierColor)

    // Pulse animation for Tier 3
    val infiniteTransition = rememberInfiniteTransition()
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.8f,
        targetValue = 1.0f,
        animationSpec = infiniteRepeatable(
            animation = tween(600),
            repeatMode = RepeatMode.Reverse
        )
    )

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(
                width = if (tier == AlertTier.TIER_3_DANGER) 3.dp else 1.dp,
                color = if (tier == AlertTier.TIER_3_DANGER) animatedColor.copy(alpha = pulseAlpha) else animatedColor,
                shape = RoundedCornerShape(16.dp)
            ),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = animatedColor.copy(alpha = 0.12f)
        )
    ) {
        Column(
            modifier = Modifier.padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = title,
                    color = animatedColor,
                    fontWeight = FontWeight.Black,
                    fontSize = 20.sp
                )
                if (tier == AlertTier.TIER_3_DANGER) {
                    Button(
                        onClick = onEmergencySosClick,
                        colors = ButtonDefaults.buttonColors(containerColor = TierDanger),
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text("EMERGENCY SOS", color = Color.White, fontWeight = FontWeight.Bold, fontSize = 12.sp)
                    }
                }
            }

            Text(
                text = subtitle,
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onSurface,
                fontWeight = FontWeight.Medium
            )

            if (distanceMeters != null) {
                Text(
                    text = "Proximity: Approximately ${distanceMeters.toInt()} meters away",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold,
                    color = animatedColor
                )
            }
        }
    }
}

@Composable
fun SpatialDistanceTrackingCard(
    distanceMeters: Double,
    tier: AlertTier
) {
    val (tierColor, rangeDesc) = when {
        distanceMeters < 250 -> TierDanger to "Critical Proximity (< 250m) • Immediate Encounter Zone"
        distanceMeters < 800 -> TierWarning to "Warning Buffer (250m - 800m) • Active Scent Corridor"
        else -> TierNotice to "Awareness Perimeter (> 800m) • Peripheral Monitoring"
    }

    val progress = ((2000.0 - distanceMeters.coerceIn(50.0, 2000.0)) / 1950.0).toFloat()

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
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Spatial Distance Tracking",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Text(
                    text = "${distanceMeters.toInt()}m",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Black,
                    color = tierColor
                )
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(8.dp)
                    .background(Color.LightGray.copy(alpha = 0.3f), RoundedCornerShape(4.dp))
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(fraction = progress.coerceIn(0.05f, 1.0f))
                        .height(8.dp)
                        .background(tierColor, RoundedCornerShape(4.dp))
                )
            }

            Text(
                text = rangeDesc,
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
            )
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
fun DetourSuggestionCard(detour: String) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.7f)),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Row(
            modifier = Modifier.padding(14.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
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
                    text = "Species-Specific Safety & Deterrent Protocol",
                    fontWeight = FontWeight.Bold,
                    fontSize = 13.sp,
                    color = MaterialTheme.colorScheme.onSurface
                )
            }

            species.forEach { sp ->
                val (name, protocol) = when (sp) {
                    WildlifeSpecies.BEAR -> "Asian Black Bear (ツキノワグマ)" to
                        "1. Make presence known with voice / bear bell.\n2. Do NOT run or climb trees; stand ground.\n3. Back away slowly facing animal.\n4. Arm capsaicin spray."
                    WildlifeSpecies.TIGER -> "Sumatran Tiger (Harimau Sumatera)" to
                        "1. Maintain upright posture; never crouch.\n2. Do NOT turn your back or run.\n3. Keep continuous eye contact.\n4. Retreat slowly to vehicle or night corral."
                    WildlifeSpecies.ELEPHANT -> "Asian Elephant (ช้างป่า)" to
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
                fontSize = 14.sp,
                lineHeight = 20.sp,
                color = MaterialTheme.colorScheme.onSurface
            )
        }
    }
}

@Composable
fun SpeciesThreatCard(
    species: List<WildlifeSpecies>,
    distanceMeters: Double?
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        )
    ) {
        Row(
            modifier = Modifier
                .padding(16.dp)
                .fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text(
                    text = "Active Target Fauna",
                    fontSize = 12.sp,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                )
                Text(
                    text = species.joinToString(", ") { it.name.replace("_", " ") },
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
            }
            if (distanceMeters != null) {
                Box(
                    modifier = Modifier
                        .background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(8.dp))
                        .padding(horizontal = 10.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "< ${distanceMeters.toInt()}m",
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Black,
                        color = MaterialTheme.colorScheme.primary
                    )
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
