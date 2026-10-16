package ai.coexistshield.wear

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.WildlifeSpecies
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import com.google.android.gms.wearable.DataEvent
import com.google.android.gms.wearable.DataEventBuffer
import com.google.android.gms.wearable.DataMapItem
import com.google.android.gms.wearable.MessageEvent
import com.google.android.gms.wearable.WearableListenerService

/**
 * Background service running on the Wear OS watch that listens for state updates
 * and immediate alert triggers from the companion mobile app.
 */
class WearDataListenerService : WearableListenerService() {

    private lateinit var hapticManager: WatchHapticManager

    override fun onCreate() {
        super.onCreate()
        hapticManager = WatchHapticManager(this)
    }

    override fun onDataChanged(dataEvents: DataEventBuffer) {
        for (event in dataEvents) {
            if (event.type == DataEvent.TYPE_CHANGED && event.dataItem.uri.path == PATH_HUD_STATE) {
                val dataMap = DataMapItem.fromDataItem(event.dataItem).dataMap
                val tierStr = dataMap.getString(KEY_TIER) ?: AlertTier.SAFE.name
                val tier = try {
                    AlertTier.valueOf(tierStr)
                } catch (_: Exception) {
                    AlertTier.SAFE
                }
                val rawDist = dataMap.getDouble(KEY_DISTANCE, -1.0)
                val distanceMeters = if (rawDist >= 0) rawDist else null
                val speciesList = dataMap.getStringArrayList(KEY_SPECIES)?.mapNotNull {
                    try {
                        WildlifeSpecies.valueOf(it)
                    } catch (_: Exception) {
                        null
                    }
                } ?: emptyList()
                val isSiren = dataMap.getBoolean(KEY_SIREN, false)

                val previousTier = WearStateRepository.hudState.value.tier
                WearStateRepository.updateState(tier, distanceMeters, speciesList, isSiren)

                // If alert tier escalated, trigger corresponding wrist haptic
                if (tier != previousTier && tier != AlertTier.SAFE) {
                    hapticManager.triggerTierHaptics(tier)
                }
            }
        }
    }

    override fun onMessageReceived(messageEvent: MessageEvent) {
        if (messageEvent.path == PATH_ALERT_TRIGGER) {
            val payload = String(messageEvent.data, Charsets.UTF_8)
            // Format: "TIER|DISTANCE|MESSAGE"
            val parts = payload.split("|")
            val tierStr = parts.getOrNull(0) ?: AlertTier.TIER_2_WARNING.name
            val tier = try {
                AlertTier.valueOf(tierStr)
            } catch (_: Exception) {
                AlertTier.TIER_2_WARNING
            }
            val distanceStr = parts.getOrNull(1)
            val distanceMeters = distanceStr?.toDoubleOrNull()
            val alertMessage = parts.getOrNull(2) ?: "Geofence Hazard Alert Detected"

            // 1. Immediately trigger distinct haptic feedback on wrist
            hapticManager.triggerTierHaptics(tier)

            // 2. Post heads-up native Wear OS notification
            postWatchAlertNotification(tier, distanceMeters, alertMessage)
        }
    }

    private fun postWatchAlertNotification(
        tier: AlertTier,
        distanceMeters: Double?,
        message: String
    ) {
        val notificationManager =
            getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID_WATCH_ALERT,
                "Wear OS Critical Hazard Alert",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "High-priority wrist notifications for dynamic geofence incursions"
                enableVibration(true)
            }
            notificationManager.createNotificationChannel(channel)
        }

        val launchIntent = Intent(this, WearMainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            launchIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val title = when (tier) {
            AlertTier.TIER_3_DANGER -> "🚨 IMMEDIATE DANGER"
            AlertTier.TIER_2_WARNING -> "⚡ TIER 2 WARNING"
            AlertTier.TIER_1_NOTICE -> "⚠️ TIER 1 NOTICE"
            AlertTier.SAFE -> "🛡️ ALL CLEAR"
        }

        val distanceText = if (distanceMeters != null) " (${distanceMeters.toInt()}m)" else ""
        val content = "$message$distanceText"

        val notification = NotificationCompat.Builder(this, CHANNEL_ID_WATCH_ALERT)
            .setSmallIcon(android.R.drawable.ic_dialog_alert)
            .setContentTitle(title)
            .setContentText(content)
            .setPriority(NotificationCompat.PRIORITY_MAX)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)
            .build()

        notificationManager.notify(NOTIFICATION_ID_WATCH_ALERT, notification)
    }

    companion object {
        const val PATH_HUD_STATE = "/hud_state"
        const val PATH_ALERT_TRIGGER = "/alert/trigger"

        const val KEY_TIER = "tier"
        const val KEY_DISTANCE = "closest_distance_meters"
        const val KEY_SPECIES = "species"
        const val KEY_SIREN = "is_siren_playing"

        const val CHANNEL_ID_WATCH_ALERT = "wear_hazard_alert_channel"
        const val NOTIFICATION_ID_WATCH_ALERT = 2001
    }
}
