package ai.coexistshield.wear

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.lifecycle.lifecycleScope
import com.google.android.gms.wearable.Wearable
import kotlinx.coroutines.launch
import kotlinx.coroutines.tasks.await

class WearMainActivity : ComponentActivity() {

    private lateinit var hapticManager: WatchHapticManager
    private val messageClient by lazy { Wearable.getMessageClient(this) }
    private val nodeClient by lazy { Wearable.getNodeClient(this) }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        hapticManager = WatchHapticManager(this)

        setContent {
            WearHudScreen(
                onSilenceClick = {
                    handleSilenceAlert()
                },
                onSosClick = {
                    handleBroadcastSos()
                }
            )
        }
    }

    private fun handleSilenceAlert() {
        // 1. Stop local wrist haptics
        hapticManager.stopHaptics()

        // 2. Update local state
        WearStateRepository.silenceSiren()

        // 3. Send dismiss message back to the phone to silence audio siren
        lifecycleScope.launch {
            try {
                val nodes = nodeClient.connectedNodes.await()
                for (node in nodes) {
                    messageClient.sendMessage(
                        node.id,
                        WearDataListenerService.PATH_ALERT_TRIGGER.replace("trigger", "dismiss"),
                        "DISMISS".toByteArray(Charsets.UTF_8)
                    ).await()
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    private fun handleBroadcastSos() {
        hapticManager.stopHaptics()
        lifecycleScope.launch {
            try {
                val nodes = nodeClient.connectedNodes.await()
                for (node in nodes) {
                    messageClient.sendMessage(
                        node.id,
                        "/alert/sos",
                        "EMERGENCY_SOS".toByteArray(Charsets.UTF_8)
                    ).await()
                }
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        hapticManager.stopHaptics()
    }
}
