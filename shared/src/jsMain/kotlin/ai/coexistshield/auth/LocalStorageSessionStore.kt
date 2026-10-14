package ai.coexistshield.auth

import kotlinx.browser.localStorage

/** Browser localStorage-backed session store. */
class LocalStorageSessionStore(private val key: String = "coexist.refreshToken") : SessionStore {
    override fun load(): String? = runCatching { localStorage.getItem(key) }.getOrNull()?.takeIf { it.isNotBlank() }
    override fun save(refreshToken: String) { runCatching { localStorage.setItem(key, refreshToken) } }
    override fun clear() { runCatching { localStorage.removeItem(key) } }
}
