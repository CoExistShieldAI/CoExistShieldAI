package ai.coexistshield.ui.map

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.ImageBitmap
import io.ktor.client.HttpClient
import io.ktor.client.request.get
import io.ktor.client.statement.readRawBytes
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import org.jetbrains.compose.resources.ExperimentalResourceApi
import org.jetbrains.compose.resources.decodeToImageBitmap
import kotlin.math.PI
import kotlin.math.asinh
import kotlin.math.tan

object TileCoordinateUtils {
    fun lonToTileX(lon: Double, zoom: Int): Double {
        return (lon + 180.0) / 360.0 * (1 shl zoom)
    }

    fun latToTileY(lat: Double, zoom: Int): Double {
        val latRad = lat.coerceIn(-85.0511, 85.0511) * PI / 180.0
        return (1.0 - asinh(tan(latRad)) / PI) / 2.0 * (1 shl zoom)
    }

    fun tileXToLon(x: Double, zoom: Int): Double {
        return x / (1 shl zoom) * 360.0 - 180.0
    }

    fun tileYToLat(y: Double, zoom: Int): Double {
        val n = PI - 2.0 * PI * y / (1 shl zoom)
        return 180.0 / PI * kotlin.math.atan(0.5 * (kotlin.math.exp(n) - kotlin.math.exp(-n)))
    }
}

class TileManager(
    private val httpClient: HttpClient = HttpClient()
) {
    private val memoryCache = mutableMapOf<String, ImageBitmap>()
    private val pendingRequests = mutableSetOf<String>()
    
    // State trigger to notify Compose when a new tile is loaded
    var cacheVersion by mutableStateOf(0)
        private set

    fun getTile(zoom: Int, x: Int, y: Int, scope: CoroutineScope): ImageBitmap? {
        val maxTile = (1 shl zoom)
        val normalizedX = (x % maxTile + maxTile) % maxTile
        if (y < 0 || y >= maxTile) return null

        val key = "$zoom/$normalizedX/$y"
        val cached = memoryCache[key]
        if (cached != null) {
            return cached
        }

        if (pendingRequests.add(key)) {
            scope.launch(Dispatchers.Default) {
                try {
                    // ESRI World Dark Gray Canvas: High-performance, clean dark HUD theme, zero watermarks, CORS-friendly
                    val url = "https://services.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Base/MapServer/tile/$zoom/$y/$normalizedX"
                    
                    val response = httpClient.get(url)
                    val bytes = response.readRawBytes()
                    @OptIn(ExperimentalResourceApi::class)
                    val bitmap = bytes.decodeToImageBitmap()
                    
                    memoryCache[key] = bitmap
                    cacheVersion++
                } catch (e: Exception) {
                    // Gracefully ignore individual tile network failures
                } finally {
                    pendingRequests.remove(key)
                }
            }
        }

        return null
    }
}
