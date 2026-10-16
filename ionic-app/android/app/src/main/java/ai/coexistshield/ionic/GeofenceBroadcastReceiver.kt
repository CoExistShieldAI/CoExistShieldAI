package ai.coexistshield.ionic

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.google.android.gms.location.Geofence
import com.google.android.gms.location.GeofencingEvent

/**
 * OS-level BroadcastReceiver that wakes up in the background when the user enters, dwells in,
 * or exits a registered wildlife hazard geofence.
 */
class GeofenceBroadcastReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        val geofencingEvent = GeofencingEvent.fromIntent(intent) ?: return
        if (geofencingEvent.hasError()) return

        val transitionType = when (geofencingEvent.geofenceTransition) {
            Geofence.GEOFENCE_TRANSITION_ENTER -> "ENTER"
            Geofence.GEOFENCE_TRANSITION_DWELL -> "DWELL"
            Geofence.GEOFENCE_TRANSITION_EXIT -> "EXIT"
            else -> return
        }

        val triggeringGeofences = geofencingEvent.triggeringGeofences ?: return
        for (fence in triggeringGeofences) {
            GeofenceBridgePlugin.instance?.emitGeofenceTransition(fence.requestId, transitionType)

            // Direct OS-to-Watch alert: Trigger wrist vibration even if Ionic WebView is sleeping
            if (transitionType == "ENTER" || transitionType == "DWELL") {
                WearSyncPlugin.broadcastUrgentAlertToWatch(
                    context = context,
                    tier = "TIER_2_WARNING",
                    distanceMeters = null,
                    message = "Entered Wildlife Geofence Zone: ${fence.requestId}"
                )
            }
        }
    }
}
