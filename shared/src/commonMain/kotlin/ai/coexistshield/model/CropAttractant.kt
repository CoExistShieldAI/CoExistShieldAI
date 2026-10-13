package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class CropType {
    SUGARCANE,
    BANANA,
    JACKFRUIT,
    MANGO,
    PALM_OIL,
    OTHER
}

@Serializable
enum class CropGrowthStage {
    VEGETATIVE,
    RIPENING,
    READY_FOR_HARVEST,
    HARVESTED_CLEARED
}

@Serializable
data class CropAttractantPlot(
    val id: String,
    val location: GeoLocation,
    val cropType: CropType,
    val growthStage: CropGrowthStage = CropGrowthStage.RIPENING,
    val attractantRiskIndex: Double = 0.0,
    val plantationName: String,
    val areaHectares: Double? = null,
    val updatedAt: String
)
