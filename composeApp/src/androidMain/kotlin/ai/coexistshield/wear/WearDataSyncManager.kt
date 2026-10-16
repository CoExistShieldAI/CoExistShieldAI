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
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import kotlinx.coroutines.tasks.await

/**
 * Manages communication from the Mobile Companion App to Wear OS devices via
 * Google Play Services Wearable Data Layer API.
 */
class WearDataSyncManager(private val context: Context) {
    private val dataClient by lazy { Wearable.getDataClient(context) }
    private val messageClient by lazy { Wearable.getMessageClient(context) }
    private val nodeClient by lazy { Wearable.getNodeClient(context) }

    companion object {
        const val PATH_HUD_STATE = "/hud_state"
        const val PATH_ALERT_TRIGGER = "/alert/trigger"
        const val PATH_ALERT_DISMISS = "/alert/dismiss"

        const val KEY_TIER = "tier"
        const val KEY_DISTANCE = "closest_distance_meters"
        const val KEY_SPECIES = "species"
        const val KEY_SIREN = "is_siren_playing"
        const val KEY_TIMESTAMP = "timestamp"

        const val CHANNEL_ID_MOBILE_HAZARD = "coexist_shield_hazard_channel"
    }

    /**
     * Synchronize latest HUD presentation state to all connected Wear OS watches.
     * Uses DataClient so state is cached and re-delivered even if the watch is briefly disconnected.
     */
    suspend fun syncHudState(
        tier: AlertTier,
        distanceMeters: Double?,
        species: List<WildlifeSpecies> = emptyList(),
        isSirenPlaying: Boolean = false
    ): Boolean {
        return try {
            val request = PutDataMapRequest.create(PATH_HUD_STATE).apply {
                dataMap.putString(KEY_TIER, tier.name)
                dataMap.putDouble(KEY_DISTANCE, distanceMeters ?: -1.0)
                dataMap.putStringArrayList(KEY_SPECIES, ArrayList(species.map { it.name }))
                dataMap.putBoolean(KEY_SIREN, isSirenPlaying)
                dataMap.putLong(KEY_TIMESTAMP, System.currentTimeMillis())
            }
            dataClient.putDataItem(request.asPutDataRequest().setUrgent()).await()
            true
        } catch (e: Exception) {
            e.printStackTrace()
            false
        }
    }

    /**
     * Broadcast an urgent alert message to all connected nodes for immediate haptic/siren trigger.
     */
    suspend fun sendImmediateAlert(
        tier: AlertTier,
        distanceMeters: Double?,
        messageText: String
    ): Boolean {
        return try {
            val connectedNodes = nodeClient.connectedNodes.await()
            val payload = "${tier.name}|$distanceMeters|$messageText".toByteArray(Charsets.UTF_8)
            for (node in connectedNodes) {
                messageClient.sendMessage(node.id, PATH_ALERT_TRIGGER, payload).await()
            }
            true
        } catch (e: Exception) {
            e.printStackTrace()
            false
        }
    }

    /**
     * Post a high-priority local notification on the Mobile Phone with setLocalOnly(true)
     * to prevent duplicate, redundant Android system mirroring on Wear OS.
     */
    fun postMobileNotificationWithMirroringSuppression(
        title: String,
        body: String,
        tier: AlertTier,
        notificationId: Int = 1001
    ) {
        val notificationManager =
            context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID_MOBILE_HAZARD,
                "CoExist Shield Hazard Notifications",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Critical wildlife encounter and geofence alerts"
            }
            notificationManager.createNotificationChannel(channel)
        }

        val launchIntent = context.packageManager.getLaunchIntentForPackage(context.packageName)?.apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            0,
            launchIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val notification = NotificationCompat.Builder(context, CHANNEL_ID_MOBILE_HAZARD)
            .setSmallIcon(android.R.drawable.ic_dialog_alert)
            .setContentTitle(title)
            .setContentText(body)
            .setPriority(NotificationCompat.PRIORITY_MAX)
            .setContentIntent(pendingIntent)
            .setAutoCancel(true)
            // ★ Critical: Suppress automatic system mirroring to Wear OS.
            // Wear OS will handle its own native, tier-vibrating notification via WearDataListenerService.
            .setLocalOnly(true)
            .build()

        notificationManager.notify(notificationId, notification)
    }
}
