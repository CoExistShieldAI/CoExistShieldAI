package ai.coexistshield.auth

/** Persists the refresh token so a session can be restored after a reload. */
interface SessionStore {
    fun load(): String?
    fun save(refreshToken: String)
    fun clear()
}

class InMemorySessionStore : SessionStore {
    private var token: String? = null
    override fun load(): String? = token
    override fun save(refreshToken: String) { token = refreshToken }
    override fun clear() { token = null }
}
