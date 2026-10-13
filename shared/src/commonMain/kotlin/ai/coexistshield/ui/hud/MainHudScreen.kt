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
import androidx.compose.material3.IconButton
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
import androidx.compose.ui.draw.alpha
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
    val activeRole: UserRole = UserRole.CITIZEN
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MainHudScreen(
    state: MainHudUiState,
    onOpenMapClick: () -> Unit = {},
    onQuickReportClick: () -> Unit = {},
    onOfficialDispatchClick: () -> Unit = {},
    onEmergencySosClick: () -> Unit = {}
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
                    ConnectionStatusIndicator(isOffline = state.isOfflineMode)
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
            // 1. Primary Alert Banner
            AlertBannerCard(
                tier = state.currentTier,
                species = state.targetSpecies,
                distanceMeters = state.closestDistanceMeters,
                onEmergencySosClick = onEmergencySosClick
            )

            // 2. AI Causal Explanation (Explainable AI Grounding)
            CausalExplanationCard(
                explanation = state.causalExplanation,
                tier = state.currentTier
            )

            // 3. Species Threat & Attractant Summary
            if (state.targetSpecies.isNotEmpty()) {
                SpeciesThreatCard(
                    species = state.targetSpecies,
                    distanceMeters = state.closestDistanceMeters
                )
            }

            Spacer(modifier = Modifier.weight(1f, fill = false))

            // 4. Primary Actions Footer
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
fun ConnectionStatusIndicator(isOffline: Boolean) {
    Row(
        modifier = Modifier.padding(end = 12.dp),
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
        Spacer(modifier = Modifier.width(4.dp))
        Text(
            text = if (isOffline) "Offline (Cached)" else "Live",
            fontSize = 12.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
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
                    text = "• Powered by Gemini",
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
