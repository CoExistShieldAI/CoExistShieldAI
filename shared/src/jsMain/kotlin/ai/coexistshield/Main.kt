package ai.coexistshield

import ai.coexistshield.auth.AuthConfig
import ai.coexistshield.auth.LocalStorageSessionStore
import ai.coexistshield.ui.App
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.window.CanvasBasedWindow
import org.jetbrains.skiko.wasm.onWasmReady

@OptIn(ExperimentalComposeUiApi::class)
fun main() {
    AuthConfig.apiKey = (js("window.COEXIST_FIREBASE_API_KEY") as? String).orEmpty()
    onWasmReady {
        CanvasBasedWindow(
            title = "CoExist Shield AI",
            canvasElementId = "ComposeTarget"
        ) {
            App(sessionStore = LocalStorageSessionStore())
        }
    }
}
