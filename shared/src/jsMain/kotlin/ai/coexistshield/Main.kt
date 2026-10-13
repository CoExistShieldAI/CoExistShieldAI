package ai.coexistshield

import ai.coexistshield.ui.App
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.window.CanvasBasedWindow
import org.jetbrains.skiko.wasm.onWasmReady

@OptIn(ExperimentalComposeUiApi::class)
fun main() {
    onWasmReady {
        CanvasBasedWindow(
            title = "CoExist Shield AI",
            canvasElementId = "ComposeTarget"
        ) {
            App()
        }
    }
}
