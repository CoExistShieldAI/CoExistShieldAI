package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class FruitType {
    PERSIMMON,
    CHESTNUT,
    APPLE,
    OTHER
}

@Serializable
enum class FruitLoadLevel {
    SPARSE,
    MODERATE,
    ABUNDANT
}

@Serializable
enum class RipenessStage {
    UNRIPE_GREEN,
    RIPE,
    OVERRIPE_FALLING
}

@Serializable
enum class PropertyContextTag {
    VACANT_HOUSE,
    ROADSIDE_UNMANAGED,
    ACTIVE_FARM_BORDER,
    PUBLIC_LAND,
    UNKNOWN
}

@Serializable
enum class TreeManagementStatus {
    UNMANAGED,
    SCHEDULED_REMOVAL,
    PRUNED,
    FELLED
}

@Serializable
data class KakiTree(
    val id: String,
    val location: GeoLocation,
    val fruitType: FruitType = FruitType.PERSIMMON,
    val fruitLoadLevel: FruitLoadLevel = FruitLoadLevel.MODERATE,
    val ripenessStage: RipenessStage = RipenessStage.RIPE,
    val isTop5PercentHazard: Boolean = false,
    val hazardScore: Double = 0.0,
    val propertyContext: PropertyContextTag = PropertyContextTag.UNKNOWN,
    val isNearSchoolRoute: Boolean = false,
    val status: TreeManagementStatus = TreeManagementStatus.UNMANAGED,
    val gcfCampaignId: String? = null,
    val photoUrls: List<String> = emptyList(),
    val reportedBy: String,
    val reportedAt: String,
    val verifiedByOfficialId: String? = null
)
