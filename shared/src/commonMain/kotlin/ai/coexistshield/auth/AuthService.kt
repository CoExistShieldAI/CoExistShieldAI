package ai.coexistshield.auth

import ai.coexistshield.model.UserRole
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.forms.submitForm
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.parameters
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlin.io.encoding.Base64
import kotlin.io.encoding.ExperimentalEncodingApi

/** Runtime auth configuration. The Firebase Web API key is injected by the platform entry point. */
object AuthConfig {
    var apiKey: String = ""
}

/** Authenticated session. */
data class AuthUser(
    val uid: String,
    val email: String,
    val role: UserRole,
    val idToken: String,
    val refreshToken: String
)

class AuthException(message: String) : Exception(message)

object AuthValidator {
    const val MIN_PASSWORD_LENGTH = 8
    private val EMAIL_REGEX = Regex("^[A-Za-z0-9._%+\\-]+@[A-Za-z0-9.\\-]+\\.[A-Za-z]{2,}$")

    fun validateEmail(email: String): String? = when {
        email.isBlank() -> "Please enter your email address"
        !EMAIL_REGEX.matches(email.trim()) -> "Invalid email address format"
        else -> null
    }

    fun validatePassword(password: String): String? = when {
        password.isEmpty() -> "Please enter a password"
        password.length < MIN_PASSWORD_LENGTH -> "Password must be at least ${MIN_PASSWORD_LENGTH} characters"
        else -> null
    }

    fun validateConfirm(password: String, confirm: String): String? =
        if (password != confirm) "Passwords do not match" else null
}

@Serializable
private data class CredentialRequest(
    val email: String,
    val password: String,
    val returnSecureToken: Boolean = true
)

@Serializable
private data class CredentialResponse(
    val idToken: String,
    val refreshToken: String,
    val localId: String,
    val email: String = ""
)

@Serializable
private data class ResetRequest(val requestType: String = "PASSWORD_RESET", val email: String)

@Serializable
private data class RefreshResponse(
    @SerialName("id_token") val idToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("user_id") val userId: String
)

@Serializable
private data class ErrorEnvelope(val error: ErrorBody? = null)

@Serializable
private data class ErrorBody(val message: String = "")

class AuthService(
    private val client: HttpClient = createClient(),
    private val apiKeyProvider: () -> String = { AuthConfig.apiKey }
) {
    suspend fun signUp(email: String, password: String): Result<AuthUser> =
        call("accounts:signUp", email, password)

    suspend fun signIn(email: String, password: String): Result<AuthUser> =
        call("accounts:signInWithPassword", email, password)

    /** Sends a password reset email. Always reports success for unknown emails (no account enumeration). */
    suspend fun sendPasswordReset(email: String): Result<Unit> {
        val apiKey = apiKeyProvider()
        if (apiKey.isBlank()) return Result.failure(AuthException(NOT_CONFIGURED))
        return try {
            val response = client.post("$BASE_URL/accounts:sendOobCode?key=$apiKey") {
                contentType(ContentType.Application.Json)
                setBody(ResetRequest(email = email.trim()))
            }
            if (response.status.isSuccess()) {
                Result.success(Unit)
            } else {
                val code = runCatching { response.body<ErrorEnvelope>().error?.message }.getOrNull().orEmpty()
                if (code.startsWith("EMAIL_NOT_FOUND")) Result.success(Unit)
                else Result.failure(AuthException(translateError(code)))
            }
        } catch (e: Exception) {
            Result.failure(AuthException(NETWORK_ERROR))
        }
    }

    /** Exchanges a refresh token for a fresh ID token (session restore / periodic renewal). */
    suspend fun refresh(refreshToken: String): Result<AuthUser> {
        val apiKey = apiKeyProvider()
        if (apiKey.isBlank()) return Result.failure(AuthException(NOT_CONFIGURED))
        return try {
            val response = client.submitForm(
                url = "$SECURE_TOKEN_URL?key=$apiKey",
                formParameters = parameters {
                    append("grant_type", "refresh_token")
                    append("refresh_token", refreshToken)
                }
            )
            if (response.status.isSuccess()) {
                val body: RefreshResponse = response.body()
                Result.success(
                    AuthUser(
                        uid = body.userId,
                        email = emailFromIdToken(body.idToken),
                        role = roleFromIdToken(body.idToken),
                        idToken = body.idToken,
                        refreshToken = body.refreshToken
                    )
                )
            } else {
                Result.failure(AuthException("Session has expired. Please sign in again."))
            }
        } catch (e: Exception) {
            Result.failure(AuthException(NETWORK_ERROR))
        }
    }

    private suspend fun call(endpoint: String, email: String, password: String): Result<AuthUser> {
        val apiKey = apiKeyProvider()
        if (apiKey.isBlank()) {
            return Result.failure(AuthException(NOT_CONFIGURED))
        }
        return try {
            val response: HttpResponse = client.post("$BASE_URL/$endpoint?key=$apiKey") {
                contentType(ContentType.Application.Json)
                setBody(CredentialRequest(email.trim(), password))
            }
            if (response.status.isSuccess()) {
                val body: CredentialResponse = response.body()
                Result.success(
                    AuthUser(
                        uid = body.localId,
                        email = body.email.ifBlank { email.trim() },
                        role = roleFromIdToken(body.idToken),
                        idToken = body.idToken,
                        refreshToken = body.refreshToken
                    )
                )
            } else {
                val code = runCatching { response.body<ErrorEnvelope>().error?.message }.getOrNull().orEmpty()
                Result.failure(AuthException(translateError(code)))
            }
        } catch (e: Exception) {
            Result.failure(AuthException(NETWORK_ERROR))
        }
    }

    companion object {
        private const val BASE_URL = "https://identitytoolkit.googleapis.com/v1"
        private const val SECURE_TOKEN_URL = "https://securetoken.googleapis.com/v1/token"
        private const val NOT_CONFIGURED = "Authentication is not configured (Firebase API key missing)"
        private const val NETWORK_ERROR = "A network error occurred. Please check your connection."

        fun createClient(): HttpClient = HttpClient {
            install(ContentNegotiation) {
                json(Json { ignoreUnknownKeys = true; isLenient = true })
            }
        }

        fun translateError(code: String): String = when {
            code.startsWith("EMAIL_EXISTS") -> "This email address is already registered"
            code.startsWith("WEAK_PASSWORD") -> "Password is too weak (must be at least 8 characters)"
            code.startsWith("INVALID_EMAIL") -> "Invalid email address format"
            code.startsWith("EMAIL_NOT_FOUND") ||
                code.startsWith("INVALID_PASSWORD") ||
                code.startsWith("INVALID_LOGIN_CREDENTIALS") -> "Invalid email address or password"
            code.startsWith("USER_DISABLED") -> "This account has been disabled"
            code.startsWith("TOO_MANY_ATTEMPTS_TRY_LATER") -> "Too many attempts. Please try again later"
            code.startsWith("OPERATION_NOT_ALLOWED") -> "Email/Password sign-in is not enabled"
            code.startsWith("CONFIGURATION_NOT_FOUND") -> "Authentication provider configuration not found"
            else -> "Authentication failed${if (code.isNotBlank()) " ($code)" else ""}"
        }

        @OptIn(ExperimentalEncodingApi::class)
        private fun claimsOf(idToken: String): JsonObject? = try {
            val payload = idToken.split(".").getOrNull(1)
            if (payload == null) null else {
                val json = Base64.UrlSafe.withPadding(Base64.PaddingOption.ABSENT_OPTIONAL)
                    .decode(payload).decodeToString()
                Json.parseToJsonElement(json) as? JsonObject
            }
        } catch (e: Exception) {
            null
        }

        fun emailFromIdToken(idToken: String): String =
            (claimsOf(idToken)?.get("email") as? JsonPrimitive)?.content.orEmpty()

        /** Reads the `role` custom claim (set by the backend/Admin SDK); defaults to CITIZEN. */
        fun roleFromIdToken(idToken: String): UserRole {
            val claim = (claimsOf(idToken)?.get("role") as? JsonPrimitive)?.content
            return UserRole.entries.firstOrNull { it.name == claim } ?: UserRole.CITIZEN
        }
    }
}
