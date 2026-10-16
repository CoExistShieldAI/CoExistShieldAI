package ai.coexistshield.ionic

import com.getcapacitor.JSArray
import com.getcapacitor.JSObject
import com.getcapacitor.Plugin
import com.getcapacitor.PluginCall
import com.getcapacitor.PluginMethod
import com.getcapacitor.annotation.CapacitorPlugin
import com.google.android.gms.wearable.MessageClient
import com.google.android.gms.wearable.MessageEvent
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.tasks.await

/**
 * Capacitor Plugin bridging Angular/Ionic state with Wear OS smartwatches via
 * Google Play Services Wearable Data Layer API.
 */
@CapacitorPlugin(name = "WearSync")
class WearSyncPlugin : Plugin(), MessageClient.OnMessageReceivedListener {

    private val scope = CoroutineScope(Dispatchers.IO)
    private val dataClient by lazy { Wearable.getDataClient(context) }
    private val messageClient by lazy { Wearable.getMessageClient(context) }
    private val nodeClient by lazy { Wearable.getNodeClient(context) }

    companion object {
        var instance: WearSyncPlugin? = null
        const val PATH_HUD_STATE = "/hud_state"
        const val PATH_ALERT_TRIGGER = "/alert/trigger"
        const val PATH_ALERT_DISMISS = "/alert/dismiss"
        const val PATH_ALERT_SOS = "/alert/sos"

        /**
         * Static helper so GeofenceBroadcastReceiver can broadcast immediate alerts
         * directly to the Wear OS watch even if the Capacitor WebView is backgrounded or killed.
         */
        fun broadcastUrgentAlertToWatch(context: android.content.Context, tier: String, distanceMeters: Double?, message: String) {
            val mClient = Wearable.getMessageClient(context)
            val nClient = Wearable.getNodeClient(context)
            CoroutineScope(Dispatchers.IO).launch {
                try {
                    val nodes = nClient.connectedNodes.await()
                    val payload = "$tier|${distanceMeters ?: -1.0}|$message".toByteArray(Charsets.UTF_8)
                    for (node in nodes) {
                        mClient.sendMessage(node.id, PATH_ALERT_TRIGGER, payload).await()
                    }
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }
        }
    }

    override fun load() {
        super.load()
        instance = this
        messageClient.addListener(this)
    }

    override fun handleOnDestroy() {
        super.handleOnDestroy()
        messageClient.removeListener(this)
        if (instance == this) instance = null
    }

    /**
     * Synchronizes current HUD state from Angular/Tailwind to connected Wear OS watch.
     */
    @PluginMethod
    fun syncHudState(call: PluginCall) {
        val tier = call.getString("tier", "SAFE") ?: "SAFE"
        val distance = if (call.hasOption("distanceMeters") && !call.isNull("distanceMeters")) {
            call.getDouble("distanceMeters")
        } else {
            -1.0
        }
        val speciesArray = call.getArray("species") ?: JSArray()
        val speciesList = ArrayList<String>()
        for (i in 0 until speciesArray.length()) {
            speciesList.add(speciesArray.optString(i))
        }
        val isSiren = call.getBoolean("isSiren", false) ?: false

        scope.launch {
            try {
                val request = PutDataMapRequest.create(PATH_HUD_STATE).apply {
                    dataMap.putString("tier", tier)
                    dataMap.putDouble("closest_distance_meters", distance)
                    dataMap.putStringArrayList("species", speciesList)
                    dataMap.putBoolean("is_siren_playing", isSiren)
                    dataMap.putLong("timestamp", System.currentTimeMillis())
                }
                dataClient.putDataItem(request.asPutDataRequest().setUrgent()).await()

                val res = JSObject().apply {
                    put("success", true)
                    put("tier", tier)
                }
                call.resolve(res)
            } catch (e: Exception) {
                call.reject("Failed to sync HUD state to Wear OS: ${e.message}", e)
            }
        }
    }

    /**
     * Sends an immediate, high-priority alert message causing the watch to vibrate instantly.
     */
    @PluginMethod
    fun sendImmediateAlert(call: PluginCall) {
        val tier = call.getString("tier", "TIER_2_WARNING") ?: "TIER_2_WARNING"
        val distance = if (call.hasOption("distanceMeters") && !call.isNull("distanceMeters")) {
            call.getDouble("distanceMeters")
        } else {
            -1.0
        }
        val message = call.getString("message", "Wildlife threat alert") ?: "Wildlife threat alert"

        scope.launch {
            try {
                val nodes = nodeClient.connectedNodes.await()
                val payload = "$tier|$distance|$message".toByteArray(Charsets.UTF_8)
                var sentCount = 0
                for (node in nodes) {
                    messageClient.sendMessage(node.id, PATH_ALERT_TRIGGER, payload).await()
                    sentCount++
                }

                val res = JSObject().apply {
                    put("success", true)
                    put("nodesCount", sentCount)
                }
                call.resolve(res)
            } catch (e: Exception) {
                call.reject("Failed to send alert to Wear OS: ${e.message}", e)
            }
        }
    }

    /**
     * Listens for reverse commands sent from the watch (e.g. user tapped "SILENCE / ACK" or "EVACUATE SOS").
     */
    override fun onMessageReceived(messageEvent: MessageEvent) {
        val path = messageEvent.path
        if (path.endsWith("dismiss") || path == PATH_ALERT_DISMISS) {
            val data = JSObject().apply {
                put("action", "DISMISS")
                put("timestamp", System.currentTimeMillis())
            }
            notifyListeners("onWatchAction", data)
        } else if (path.endsWith("sos") || path == PATH_ALERT_SOS) {
            val data = JSObject().apply {
                put("action", "SOS")
                put("timestamp", System.currentTimeMillis())
            }
            notifyListeners("onWatchAction", data)
        }
    }
}
