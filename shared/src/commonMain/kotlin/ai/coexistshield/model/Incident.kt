package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class ReportCategory {
    WILDLIFE_SIGHTING,
    PHYSICAL_TRACE,
    ROADKILL,
    LIVESTOCK_DEPREDATION,
    ATTRACTANT_FRUIT_TREE
}

@Serializable
enum class VerificationStatus {
    PENDING,
    VERIFIED,
    REJECTED
}

@Serializable
data class GeminiTriageResult(
    val detectedLabel: String? = null,
    val commonName: String? = null,
    val confidence: Double = 0.0,
    val forensicTags: List<String> = emptyList(),
    val decayState: String? = null,
    val fruitLoadLevel: String? = null,
    val notes: String? = null
)

@Serializable
data class Incident(
    val id: String,
    val category: ReportCategory,
    val species: WildlifeSpecies = WildlifeSpecies.UNKNOWN,
    val location: GeoLocation,
    val timestamp: String, // ISO-8601 string
    val region: String = "JAPAN", // "JAPAN" | "APAC"
    val description: String? = null,
    val photoUrls: List<String> = emptyList(),
    val geminiTriageResult: GeminiTriageResult? = null,
    val reporterId: String,
    val reporterRole: UserRole = UserRole.CITIZEN,
    val verificationStatus: VerificationStatus = VerificationStatus.PENDING
)
