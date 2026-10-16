package ai.coexistshield.wear

import ai.coexistshield.model.AlertTier
import android.content.Context
import android.os.Build
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager

/**
 * Manages tactile haptics tailored specifically to alert tiers on Wear OS wrist devices.
 * Designed so users can instantly discern threat severity by touch alone, even without looking.
 */
class WatchHapticManager(private val context: Context) {

    private val vibrator: Vibrator? = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
        val vibratorManager =
            context.getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as? VibratorManager
        vibratorManager?.defaultVibrator
    } else {
        @Suppress("DEPRECATION")
        context.getSystemService(Context.VIBRATOR_SERVICE) as? Vibrator
    }

    /**
     * Trigger distinctive vibration waveform based on the threat tier.
     */
    fun triggerTierHaptics(tier: AlertTier) {
        val vib = vibrator ?: return
        if (!vib.hasVibrator()) return

        stopHaptics() // Cancel any ongoing vibration first

        when (tier) {
            AlertTier.SAFE -> {
                // Tier 0 Safe: Subtle confirming tick (50ms)
                val effect = VibrationEffect.createOneShot(50, VibrationEffect.DEFAULT_AMPLITUDE)
                vib.vibrate(effect)
            }

            AlertTier.TIER_1_NOTICE -> {
                // Tier 1 Notice: Double Pulse (Tap-Tap)
                // Timings: 0ms -> 150ms vibration -> 150ms pause -> 150ms vibration
                val timings = longArrayOf(0, 150, 150, 150)
                val amplitudes = intArrayOf(0, 130, 0, 130)
                val effect = VibrationEffect.createWaveform(timings, amplitudes, -1) // No loop
                vib.vibrate(effect)
            }

            AlertTier.TIER_2_WARNING -> {
                // Tier 2 Warning: Escalating 3-Pulse Ramp (Short -> Medium -> Long & Strong)
                // Timings: 0ms -> 150ms (mild) -> 100ms pause -> 250ms (medium) -> 100ms pause -> 420ms (strong)
                val timings = longArrayOf(0, 150, 100, 250, 100, 420)
                val amplitudes = intArrayOf(0, 150, 0, 200, 0, 255)
                val effect = VibrationEffect.createWaveform(timings, amplitudes, -1) // No loop
                vib.vibrate(effect)
            }

            AlertTier.TIER_3_DANGER -> {
                // Tier 3 Danger: Aggressive SOS Emergency Siren Cycle
                // Continuous high-amplitude emergency pulse looping until stopped by user acknowledgment
                val timings = longArrayOf(0, 400, 120, 400, 120, 400, 350)
                val amplitudes = intArrayOf(0, 255, 0, 255, 0, 255, 0)
                val effect = VibrationEffect.createWaveform(timings, amplitudes, 1) // repeatIndex = 1 (Loop)
                vib.vibrate(effect)
            }
        }
    }

    /**
     * Stop all active vibrations immediately (e.g. when user acknowledges alert or mutes siren).
     */
    fun stopHaptics() {
        vibrator?.cancel()
    }
}
