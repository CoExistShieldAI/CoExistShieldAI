package ai.coexistshield.auth

import ai.coexistshield.model.UserRole
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
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
        email.isBlank() -> "メールアドレスを入力してください"
        !EMAIL_REGEX.matches(email.trim()) -> "メールアドレスの形式が正しくありません"
        else -> null
    }

    fun validatePassword(password: String): String? = when {
        password.isEmpty() -> "パスワードを入力してください"
        password.length < MIN_PASSWORD_LENGTH -> "パスワードは${MIN_PASSWORD_LENGTH}文字以上で入力してください"
        else -> null
    }

    fun validateConfirm(password: String, confirm: String): String? =
        if (password != confirm) "パスワードが一致しません" else null
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

    private suspend fun call(endpoint: String, email: String, password: String): Result<AuthUser> {
        val apiKey = apiKeyProvider()
        if (apiKey.isBlank()) {
            return Result.failure(AuthException("認証が未設定です (Firebase API キーを設定してください)"))
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
            Result.failure(AuthException("ネットワークエラーが発生しました。接続を確認してください"))
        }
    }

    companion object {
        private const val BASE_URL = "https://identitytoolkit.googleapis.com/v1"

        fun createClient(): HttpClient = HttpClient {
            install(ContentNegotiation) {
                json(Json { ignoreUnknownKeys = true; isLenient = true })
            }
        }

        fun translateError(code: String): String = when {
            code.startsWith("EMAIL_EXISTS") -> "このメールアドレスは既に登録されています"
            code.startsWith("WEAK_PASSWORD") -> "パスワードが弱すぎます (8文字以上にしてください)"
            code.startsWith("INVALID_EMAIL") -> "メールアドレスの形式が正しくありません"
            code.startsWith("EMAIL_NOT_FOUND") ||
                code.startsWith("INVALID_PASSWORD") ||
                code.startsWith("INVALID_LOGIN_CREDENTIALS") -> "メールアドレスまたはパスワードが正しくありません"
            code.startsWith("USER_DISABLED") -> "このアカウントは無効化されています"
            code.startsWith("TOO_MANY_ATTEMPTS_TRY_LATER") -> "試行回数が多すぎます。しばらくしてから再試行してください"
            code.startsWith("OPERATION_NOT_ALLOWED") -> "メール/パスワード認証が有効化されていません"
            else -> "認証に失敗しました${if (code.isNotBlank()) " ($code)" else ""}"
        }

        /** Reads the `role` custom claim (set by the backend/Admin SDK); defaults to CITIZEN. */
        @OptIn(ExperimentalEncodingApi::class)
        fun roleFromIdToken(idToken: String): UserRole {
            return try {
                val payload = idToken.split(".").getOrNull(1) ?: return UserRole.CITIZEN
                val json = Base64.UrlSafe.withPadding(Base64.PaddingOption.ABSENT_OPTIONAL)
                    .decode(payload).decodeToString()
                val claim = (Json.parseToJsonElement(json) as? JsonObject)?.get("role") as? JsonPrimitive
                UserRole.entries.firstOrNull { it.name == claim?.content } ?: UserRole.CITIZEN
            } catch (e: Exception) {
                UserRole.CITIZEN
            }
        }
    }
}
