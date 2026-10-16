package ai.coexistshield.ui.hud

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.WildlifeSpecies
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

// Modern tactical HUD color palette aligned with ionic-app
val TacticalSafe = Color(0xFF10B981)      // Emerald 500
val TacticalNotice = Color(0xFF38BDF8)    // Sky 400 (Vibrant High-Tech Cyan)
val TacticalWarning = Color(0xFFF59E0B)   // Amber 500
val TacticalDanger = Color(0xFFEF4444)    // Red 500

val MeterSlate400 = Color(0xFF94A3B8)
val MeterSlate100 = Color(0xFFF1F5F9)

/**
 * High-precision, modern tactical HUD gauge matching the ionic-app presentation:
 * - Ambient radial glow with threat pulse
 * - Outer track with 18% opacity
 * - Dynamic proximity arc with round cap
 * - Inner dashed tactical radar ring (stroke-dasharray 4 8 equivalent)
 * - Focused, high-contrast central typography
 */
@Composable
fun ModernTacticalMeter(
    tier: AlertTier,
    distanceMeters: Double?,
    species: List<WildlifeSpecies> = emptyList(),
    modifier: Modifier = Modifier,
    meterSize: Dp = 220.dp
) {
    val (targetColor, tierLabel, tierIcon) = when (tier) {
        AlertTier.SAFE -> Triple(TacticalSafe, "SAFE ZONE", "🛡️")
        AlertTier.TIER_1_NOTICE -> Triple(TacticalNotice, "TIER 1 NOTICE", "⚠️")
        AlertTier.TIER_2_WARNING -> Triple(TacticalWarning, "TIER 2 WARNING", "⚡")
        AlertTier.TIER_3_DANGER -> Triple(TacticalDanger, "IMMEDIATE DANGER", "🚨")
    }

    val animatedColor by animateColorAsState(
        targetValue = targetColor,
        animationSpec = tween(500)
    )

    // Dynamic proximity sweep angle calculation (matching ionic-app getArcOffset)
    // SAFE: Full 360 degree continuous perimeter
    // Other tiers: 50m proximity maps to ~92% fill, 2000m maps to ~20% fill
    val targetSweepAngle = remember(tier, distanceMeters) {
        if (tier == AlertTier.SAFE) {
            360f
        } else if (distanceMeters == null) {
            360f * 0.4f
        } else {
            val clamped = distanceMeters.coerceIn(50.0, 2000.0)
            val ratio = ((2000.0 - clamped) / 1950.0).toFloat()
            val fillRatio = ratio.coerceIn(0.2f, 0.92f)
            360f * fillRatio
        }
    }

    val animatedSweep = remember { Animatable(targetSweepAngle) }
    LaunchedEffect(targetSweepAngle) {
        animatedSweep.animateTo(
            targetValue = targetSweepAngle,
            animationSpec = tween(durationMillis = 600, easing = FastOutSlowInEasing)
        )
    }

    // Threat heartbeat pulse for Tier 3 Danger or active alert
    val infiniteTransition = rememberInfiniteTransition()
    val dangerPulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.14f,
        targetValue = if (tier == AlertTier.TIER_3_DANGER) 0.40f else 0.22f,
        animationSpec = infiniteRepeatable(
            animation = tween(if (tier == AlertTier.TIER_3_DANGER) 450 else 1200),
            repeatMode = RepeatMode.Reverse
        )
    )

    Box(
        modifier = modifier.size(meterSize),
        contentAlignment = Alignment.Center
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            val strokeWidthPx = (meterSize.value * 0.045f).coerceAtLeast(8f).dp.toPx()
            val outerRadius = (size.minDimension - strokeWidthPx) / 2f
            val innerRadius = outerRadius * 0.83f // Matches ionic SVG r=78 / r=94 ratio
            val centerOffset = Offset(size.width / 2f, size.height / 2f)

            // 1. Ambient Radial Glow
            drawCircle(
                brush = Brush.radialGradient(
                    colors = listOf(
                        animatedColor.copy(alpha = dangerPulseAlpha),
                        animatedColor.copy(alpha = 0.04f),
                        Color.Transparent
                    ),
                    center = centerOffset,
                    radius = size.width / 1.55f
                ),
                radius = outerRadius
            )

            // 2. Outer Track
            drawCircle(
                color = animatedColor.copy(alpha = 0.18f),
                radius = outerRadius,
                style = Stroke(width = strokeWidthPx)
            )

            // 3. Dynamic Proximity Arc (12 o'clock start)
            drawArc(
                color = animatedColor,
                startAngle = -90f,
                sweepAngle = animatedSweep.value,
                useCenter = false,
                style = Stroke(
                    width = strokeWidthPx,
                    cap = if (tier == AlertTier.SAFE) StrokeCap.Square else StrokeCap.Round
                )
            )

            // 4. Inner Dashed Tactical Ring
            val dashOn = (meterSize.value * 0.018f).coerceAtLeast(3f).dp.toPx()
            val dashOff = (meterSize.value * 0.036f).coerceAtLeast(6f).dp.toPx()
            val dashIntervals = floatArrayOf(dashOn, dashOff)
            drawCircle(
                color = animatedColor.copy(alpha = 0.35f),
                radius = innerRadius,
                style = Stroke(
                    width = 2.dp.toPx(),
                    pathEffect = PathEffect.dashPathEffect(dashIntervals, 0f)
                )
            )
        }

        // 5. Central Tactical Information Orb
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(horizontal = 20.dp)
        ) {
            Text(
                text = tierIcon,
                fontSize = if (meterSize < 200.dp) 26.sp else 34.sp
            )

            Text(
                text = tierLabel,
                fontWeight = FontWeight.Black,
                fontSize = if (meterSize < 200.dp) 11.sp else 13.sp,
                letterSpacing = 1.sp,
                color = animatedColor,
                textAlign = TextAlign.Center
            )

            if (distanceMeters != null) {
                Text(
                    text = "${distanceMeters.toInt()}m",
                    fontWeight = FontWeight.Black,
                    fontSize = if (meterSize < 200.dp) 20.sp else 26.sp,
                    color = MeterSlate100
                )
                Text(
                    text = "ESTIMATED PROXIMITY",
                    fontWeight = FontWeight.Bold,
                    fontSize = if (meterSize < 200.dp) 7.sp else 9.sp,
                    letterSpacing = 1.sp,
                    color = MeterSlate400
                )
            } else {
                Text(
                    text = "ALL SECURE",
                    fontWeight = FontWeight.ExtraBold,
                    fontSize = if (meterSize < 200.dp) 14.sp else 18.sp,
                    color = MeterSlate100
                )
                Text(
                    text = "CONTINUOUS SENSING",
                    fontWeight = FontWeight.Bold,
                    fontSize = if (meterSize < 200.dp) 7.sp else 9.sp,
                    letterSpacing = 1.sp,
                    color = MeterSlate400
                )
            }

            if (species.isNotEmpty() && meterSize >= 200.dp) {
                Text(
                    text = species.joinToString(" • ") { it.name },
                    fontWeight = FontWeight.Bold,
                    fontSize = 11.sp,
                    color = animatedColor,
                    maxLines = 1,
                    modifier = Modifier.padding(top = 2.dp)
                )
            }
        }
    }
}
