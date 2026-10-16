package ai.coexistshield.ionic

import android.os.Bundle
import com.getcapacitor.BridgeActivity

class MainActivity : BridgeActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        registerPlugin(GeofenceBridgePlugin::class.java)
        registerPlugin(WearSyncPlugin::class.java)
        super.onCreate(savedInstanceState)
    }
}
