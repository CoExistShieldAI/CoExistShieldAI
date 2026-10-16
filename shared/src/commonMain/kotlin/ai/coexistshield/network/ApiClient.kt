package ai.coexistshield.network

import io.ktor.client.HttpClient
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.request.header
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

object ApiClient {
    private const val DEFAULT_BASE_URL = "https://coexistshield-backend-4cerxfo4wq-as.a.run.app"
    private var authToken: String? = null

    fun setAuthToken(token: String?) {
        authToken = token
    }

    fun createHttpClient(
        baseUrl: String = DEFAULT_BASE_URL,
        customClient: HttpClient? = null
    ): HttpClient {
        return customClient ?: HttpClient {
            install(ContentNegotiation) {
                json(Json {
                    prettyPrint = true
                    isLenient = true
                    ignoreUnknownKeys = true
                })
            }
            defaultRequest {
                url(baseUrl)
                contentType(ContentType.Application.Json)
                authToken?.let { token ->
                    header("Authorization", "Bearer $token")
                }
            }
        }
    }
}
