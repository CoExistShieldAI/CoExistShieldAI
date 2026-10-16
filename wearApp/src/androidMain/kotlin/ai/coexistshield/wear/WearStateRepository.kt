package ai.coexistshield.wear

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.WildlifeSpecies
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class WearHudState(
    val tier: AlertTier = AlertTier.SAFE,
    val closestDistanceMeters: Double? = null,
    val targetSpecies: List<WildlifeSpecies> = emptyList(),
    val isSirenPlaying: Boolean = false,
    val lastUpdatedMillis: Long = System.currentTimeMillis()
)

/**
 * Singleton state holder for the Wear OS companion app.
 */
object WearStateRepository {
    private val _hudState = MutableStateFlow(WearHudState())
    val hudState: StateFlow<WearHudState> = _hudState.asStateFlow()

    fun updateState(
        tier: AlertTier,
        distanceMeters: Double?,
        species: List<WildlifeSpecies>,
        isSirenPlaying: Boolean
    ) {
        _hudState.value = WearHudState(
            tier = tier,
            closestDistanceMeters = distanceMeters,
            targetSpecies = species,
            isSirenPlaying = isSirenPlaying,
            lastUpdatedMillis = System.currentTimeMillis()
        )
    }

    fun silenceSiren() {
        _hudState.value = _hudState.value.copy(isSirenPlaying = false)
    }
}
