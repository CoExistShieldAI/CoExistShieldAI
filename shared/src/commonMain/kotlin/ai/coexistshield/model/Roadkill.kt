package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class DecayState {
    FRESH,
    BLOATED,
    DECOMPOSING,
    SKELETAL
}

@Serializable
enum class ClearanceStatus {
    REPORTED,
    DISPATCHED,
    CLEARED
}

@Serializable
data class RoadkillRecord(
    val id: String,
    val incidentId: String,
    val species: WildlifeSpecies,
    val location: GeoLocation,
    val decayState: DecayState = DecayState.FRESH,
    val attractantThreatScore: Double = 0.0,
    val clearanceStatus: ClearanceStatus = ClearanceStatus.REPORTED,
    val assignedOfficialId: String? = null,
    val photoUrls: List<String> = emptyList(),
    val reportedAt: String,
    val clearedAt: String? = null
)
