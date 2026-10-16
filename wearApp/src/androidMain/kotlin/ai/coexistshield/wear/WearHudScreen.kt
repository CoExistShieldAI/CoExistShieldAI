package ai.coexistshield.wear

import ai.coexistshield.model.AlertTier
import ai.coexistshield.ui.hud.ModernTacticalMeter
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.ButtonDefaults
import androidx.wear.compose.material.CompactButton
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText

/**
 * Fullscreen tactical Wear OS HUD screen rendering the high-precision ModernTacticalMeter
 * perfectly fitted to round watch faces.
 */
@Composable
fun WearHudScreen(
    onSilenceClick: () -> Unit = {},
    onSosClick: () -> Unit = {}
) {
    val state by WearStateRepository.hudState.collectAsState()
    var showActionOverlay by remember { mutableStateOf(false) }

    Scaffold(
        timeText = {
            if (!showActionOverlay) {
                TimeText()
            }
        },
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .clickable {
                    // Tap watch face to toggle emergency tactical action overlay
                    if (state.tier != AlertTier.SAFE || state.isSirenPlaying) {
                        showActionOverlay = !showActionOverlay
                    }
                },
            contentAlignment = Alignment.Center
        ) {
            // Main Tactical Meter (Fitted for standard 1.2 - 1.4" round watch bezel)
            ModernTacticalMeter(
                tier = state.tier,
                distanceMeters = state.closestDistanceMeters,
                species = state.targetSpecies,
                meterSize = 186.dp
            )

            // Emergency Action Overlay (When tapped or during Tier 3 Danger)
            if (showActionOverlay || state.isSirenPlaying) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(Color.Black.copy(alpha = 0.82f)),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                        modifier = Modifier.padding(16.dp)
                    ) {
                        Text(
                            text = if (state.isSirenPlaying) "🚨 SIREN ACTIVE" else "TACTICAL ACTIONS",
                            fontWeight = FontWeight.Black,
                            fontSize = 11.sp,
                            color = Color(0xFFEF4444),
                            textAlign = TextAlign.Center
                        )

                        // Mute / Acknowledge Button
                        CompactButton(
                            onClick = {
                                onSilenceClick()
                                showActionOverlay = false
                            },
                            colors = ButtonDefaults.buttonColors(
                                backgroundColor = Color(0xFF1E293B)
                            ),
                            modifier = Modifier.fillMaxWidth(0.75f)
                        ) {
                            Text(
                                text = "SILENCE / ACK",
                                fontWeight = FontWeight.Bold,
                                fontSize = 10.sp,
                                color = Color.White
                            )
                        }

                        // SOS Evacuation Dispatch Button
                        if (state.tier == AlertTier.TIER_3_DANGER) {
                            CompactButton(
                                onClick = {
                                    onSosClick()
                                    showActionOverlay = false
                                },
                                colors = ButtonDefaults.buttonColors(
                                    backgroundColor = Color(0xFFDC2626)
                                ),
                                modifier = Modifier.fillMaxWidth(0.75f)
                            ) {
                                Text(
                                    text = "EVACUATE SOS",
                                    fontWeight = FontWeight.Black,
                                    fontSize = 10.sp,
                                    color = Color.White
                                )
                            }
                        }

                        // Dismiss overlay
                        Text(
                            text = "Tap to close",
                            fontSize = 9.sp,
                            color = Color.Gray,
                            modifier = Modifier.padding(top = 4.dp)
                        )
                    }
                }
            }
        }
    }
}
