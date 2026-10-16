package ai.coexistshield.auth

import ai.coexistshield.model.UserRole
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import kotlin.io.encoding.Base64
import kotlin.io.encoding.ExperimentalEncodingApi
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

@OptIn(ExperimentalEncodingApi::class)
class AuthServiceTest {

    private fun jwt(payload: String): String =
        "h." + Base64.UrlSafe.withPadding(Base64.PaddingOption.ABSENT).encode(payload.encodeToByteArray()) + ".s"

    private fun service(body: String, status: HttpStatusCode = HttpStatusCode.OK, key: String = "k"): AuthService {
        val engine = MockEngine {
            respond(body, status, headersOf(HttpHeaders.ContentType, ContentType.Application.Json.toString()))
        }
        val client = HttpClient(engine) {
            install(ContentNegotiation) { json(Json { ignoreUnknownKeys = true }) }
        }
        return AuthService(client) { key }
    }

    @Test
    fun validatorRejectsBadInput() {
        assertNotNull(AuthValidator.validateEmail(""))
        assertNotNull(AuthValidator.validateEmail("not-an-email"))
        assertNull(AuthValidator.validateEmail("a@b.co"))
        assertNotNull(AuthValidator.validatePassword("short"))
        assertNull(AuthValidator.validatePassword("longenough"))
        assertNotNull(AuthValidator.validateConfirm("a", "b"))
        assertNull(AuthValidator.validateConfirm("a", "a"))
    }

    @Test
    fun signInSucceedsAndReadsRoleClaim() = runTest {
        val token = jwt("""{"role":"OFFICIAL"}""")
        val res = service("""{"idToken":"$token","refreshToken":"r","localId":"u1","email":"a@b.co"}""")
            .signIn("a@b.co", "password1")
        val user = res.getOrThrow()
        assertEquals("u1", user.uid)
        assertEquals(UserRole.OFFICIAL, user.role)
    }

    @Test
    fun missingOrInvalidRoleDefaultsToCitizen() {
        assertEquals(UserRole.CITIZEN, AuthService.roleFromIdToken(jwt("""{"uid":"x"}""")))
        assertEquals(UserRole.CITIZEN, AuthService.roleFromIdToken(jwt("""{"role":"ROOT"}""")))
        assertEquals(UserRole.CITIZEN, AuthService.roleFromIdToken("garbage"))
    }

    @Test
    fun signUpMapsServerErrorToMessage() = runTest {
        val res = service("""{"error":{"message":"EMAIL_EXISTS"}}""", HttpStatusCode.BadRequest)
            .signUp("a@b.co", "password1")
        assertTrue(res.isFailure)
        assertEquals("This email address is already registered", res.exceptionOrNull()?.message)
    }

    @Test
    fun blankApiKeyFailsWithoutNetworkCall() = runTest {
        val res = service("{}", key = "").signIn("a@b.co", "password1")
        assertTrue(res.isFailure)
        assertTrue(res.exceptionOrNull()?.message!!.contains("not configured"))
    }

    @Test
    fun refreshReturnsUserWithEmailAndRoleFromToken() = runTest {
        val token = jwt("""{"role":"ADMIN","email":"x@y.zz"}""")
        val res = service("""{"id_token":"$token","refresh_token":"r2","user_id":"u9"}""").refresh("r1")
        val user = res.getOrThrow()
        assertEquals("u9", user.uid)
        assertEquals("x@y.zz", user.email)
        assertEquals(UserRole.ADMIN, user.role)
        assertEquals("r2", user.refreshToken)
    }

    @Test
    fun refreshFailureIsReported() = runTest {
        val res = service("""{"error":{"message":"TOKEN_EXPIRED"}}""", HttpStatusCode.BadRequest).refresh("r1")
        assertTrue(res.isFailure)
    }

    @Test
    fun passwordResetSucceedsAndHidesUnknownEmail() = runTest {
        assertTrue(service("""{"email":"a@b.co"}""").sendPasswordReset("a@b.co").isSuccess)
        assertTrue(
            service("""{"error":{"message":"EMAIL_NOT_FOUND"}}""", HttpStatusCode.BadRequest)
                .sendPasswordReset("nobody@b.co").isSuccess
        )
        assertTrue(
            service("""{"error":{"message":"INVALID_EMAIL"}}""", HttpStatusCode.BadRequest)
                .sendPasswordReset("bad").isFailure
        )
        assertTrue(service("{}", key = "").sendPasswordReset("a@b.co").isFailure)
    }

    @Test
    fun inMemorySessionStoreRoundTrips() {
        val store = InMemorySessionStore()
        assertNull(store.load())
        store.save("tok")
        assertEquals("tok", store.load())
        store.clear()
        assertNull(store.load())
    }
}
