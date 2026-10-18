package ai.coexistshield.ui.report

import ai.coexistshield.edge.EdgeCameraMode
import ai.coexistshield.edge.EdgeVisionAnalysisResult
import ai.coexistshield.edge.EdgeVisionEngine
import ai.coexistshield.model.DecayState
import ai.coexistshield.model.FruitLoadLevel
import ai.coexistshield.model.FruitType
import ai.coexistshield.model.GeminiTriageResult
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.PropertyContextTag
import ai.coexistshield.model.ReportCategory
import ai.coexistshield.model.RipenessStage
import ai.coexistshield.model.UserRole
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.ui.theme.TierDanger
import ai.coexistshield.ui.theme.TierNotice
import ai.coexistshield.ui.theme.TierSafe
import ai.coexistshield.ui.theme.TierWarning
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

data class IncidentReportUiState(
    val photoCount: Int = 1,
    val location: GeoLocation = GeoLocation(39.702, 141.152),
    val selectedCategory: ReportCategory = ReportCategory.ATTRACTANT_FRUIT_TREE,
    val selectedSpecies: WildlifeSpecies = WildlifeSpecies.BEAR,
    val selectedTraceType: String = "CLAW_MARK_BARK_STRIPPING",
    val selectedDecayState: DecayState = DecayState.BLOATED,
    val selectedPropertyTag: PropertyContextTag = PropertyContextTag.VACANT_HOUSE,
    val isNearSchoolRoute: Boolean = true,
    val geminiTriageResult: GeminiTriageResult? = GeminiTriageResult(
        detectedLabel = "Diospyros kaki (Persimmon)",
        commonName = "Abandoned Persimmon Tree",
        confidence = 0.96,
        forensicTags = listOf("abundant_fruit", "overripe_falling", "bear_attractant"),
        fruitLoadLevel = "ABUNDANT",
        notes = "High-calorie bear attractant cluster detected within 500m of residential structure."
    ),
    val edgeVisionResult: EdgeVisionAnalysisResult = EdgeVisionEngine.analyzeFrame(),
    val isAnalyzing: Boolean = false,
    val isOfflineMode: Boolean = false,
    val offlineQueueCount: Int = 0,
    val isQueuedOffline: Boolean = false,
    val isSubmitted: Boolean = false,
    val userRole: UserRole = UserRole.CITIZEN
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun IncidentReportScreen(
    state: IncidentReportUiState,
    onBackClick: () -> Unit = {},
    onCategoryChange: (ReportCategory) -> Unit = {},
    onSpeciesChange: (WildlifeSpecies) -> Unit = {},
    onTraceTypeChange: (String) -> Unit = {},
    onDecayStateChange: (DecayState) -> Unit = {},
    onPropertyTagChange: (PropertyContextTag) -> Unit = {},
    onNearSchoolRouteChange: (Boolean) -> Unit = {},
    onToggleOfflineMode: (Boolean) -> Unit = {},
    onRetakePhotoClick: () -> Unit = {},
    onSyncOfflineQueueClick: () -> Unit = {},
    onSubmitReportClick: () -> Unit = {}
) {
    var currentStep by remember { mutableStateOf(1) }
    val activeEdgeMode = when (state.selectedCategory) {
        ReportCategory.ROADKILL, ReportCategory.LIVESTOCK_DEPREDATION -> EdgeCameraMode.KILL_SITE
        ReportCategory.ATTRACTANT_FRUIT_TREE -> EdgeCameraMode.ATTRACTANT
        else -> EdgeCameraMode.SIGHTING
    }
    val liveEdgeVision = remember(activeEdgeMode, state.selectedSpecies, state.location) {
        EdgeVisionEngine.analyzeFrame(
            mode = activeEdgeMode,
            speciesHint = state.selectedSpecies,
            location = state.location
        )
    }

    Scaffold(
        topBar = {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.surface)
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                "📸 Field Incident Report",
                                fontSize = 17.sp,
                                fontWeight = FontWeight.Bold
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Box(
                                modifier = Modifier
                                    .background(
                                        MaterialTheme.colorScheme.primary.copy(alpha = 0.15f),
                                        RoundedCornerShape(12.dp)
                                    )
                                    .padding(horizontal = 8.dp, vertical = 2.dp)
                            ) {
                                Text(
                                    text = "Step $currentStep of 3",
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = MaterialTheme.colorScheme.primary
                                )
                            }
                        }
                        Text(
                            "Guided Gemini 3.8 Flash Multimodal Verification",
                            fontSize = 11.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                    OutlinedButton(
                        onClick = onBackClick,
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text("✕ Close", fontSize = 12.sp)
                    }
                }

                // 3-Step Visual Progress Stepper
                WizardStepperBar(
                    currentStep = currentStep,
                    onStepClick = { currentStep = it }
                )
            }
        },
        bottomBar = {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.surface)
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                if (currentStep > 1) {
                    OutlinedButton(
                        onClick = { currentStep -= 1 },
                        modifier = Modifier.height(48.dp),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Text("← Back", fontSize = 13.sp, fontWeight = FontWeight.Bold)
                    }
                }

                if (currentStep < 3) {
                    Button(
                        onClick = { currentStep += 1 },
                        modifier = Modifier
                            .weight(1f)
                            .height(48.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = MaterialTheme.colorScheme.primary
                        )
                    ) {
                        Text(
                            text = if (currentStep == 1) "Next: Photo & Scene Details →" else "Next: AI Triage & Review →",
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                } else {
                    Button(
                        onClick = {
                            onSubmitReportClick()
                            currentStep = 1
                        },
                        modifier = Modifier
                            .weight(1f)
                            .height(48.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (state.isOfflineMode) TierWarning
                            else if (state.selectedCategory == ReportCategory.LIVESTOCK_DEPREDATION || state.selectedCategory == ReportCategory.ROADKILL)
                                TierDanger else TierSafe
                        )
                    ) {
                        Text(
                            text = if (state.isOfflineMode) "📦 Save to Offline Queue"
                            else "✨ Submit Verified Report",
                            fontSize = 14.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                }
            }
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Offline Pending Queue Banner (always shown when pending reports exist)
            if (state.offlineQueueCount > 0) {
                OfflineQueueBanner(
                    queueCount = state.offlineQueueCount,
                    onSyncClick = onSyncOfflineQueueClick
                )
            }

            when (currentStep) {
                1 -> {
                    // STEP 1: Visual Category Cards
                    CategorySelectorSection(
                        selectedCategory = state.selectedCategory,
                        onCategoryChange = onCategoryChange
                    )
                }

                2 -> {
                    // STEP 2: Selected Category Pill + Photo & GPS + Scene Context
                    SelectedCategoryHeaderBar(
                        selectedCategory = state.selectedCategory,
                        onChangeClick = { currentStep = 1 }
                    )

                    PhotoMetadataCard(
                        photoCount = state.photoCount,
                        location = state.location,
                        edgeVision = liveEdgeVision,
                        onAnalyzeClick = { currentStep = 3 },
                        onRetakePhotoClick = onRetakePhotoClick
                    )

                    when (state.selectedCategory) {
                        ReportCategory.WILDLIFE_SIGHTING -> {
                            WildlifeSightingSection(
                                selectedSpecies = state.selectedSpecies,
                                onSpeciesChange = onSpeciesChange
                            )
                        }
                        ReportCategory.PHYSICAL_TRACE -> {
                            PhysicalTraceContextSection(
                                selectedTraceType = state.selectedTraceType,
                                onTraceTypeChange = onTraceTypeChange
                            )
                        }
                        ReportCategory.ROADKILL -> {
                            RoadkillContextSection(
                                selectedSpecies = state.selectedSpecies,
                                selectedDecayState = state.selectedDecayState,
                                onSpeciesChange = onSpeciesChange,
                                onDecayStateChange = onDecayStateChange
                            )
                        }
                        ReportCategory.ATTRACTANT_FRUIT_TREE -> {
                            FruitTreeContextSection(
                                selectedTag = state.selectedPropertyTag,
                                isNearSchoolRoute = state.isNearSchoolRoute,
                                onPropertyTagChange = onPropertyTagChange,
                                onNearSchoolRouteChange = onNearSchoolRouteChange
                            )
                        }
                        ReportCategory.LIVESTOCK_DEPREDATION -> {
                            LivestockDepredationSection(
                                selectedSpecies = state.selectedSpecies,
                                onSpeciesChange = onSpeciesChange
                            )
                        }
                    }
                }

                3 -> {
                    // STEP 3: Predator Caching Risk + YOLOv8n/Gemini Triage + SHA-256 Evidence + Summary
                    if (liveEdgeVision.cachingEvaluation.cachingDetected) {
                        PredatorCachingRiskBanner(edgeVision = liveEdgeVision)
                    }

                    GeminiTriageCard(
                        category = state.selectedCategory,
                        result = state.geminiTriageResult,
                        isAnalyzing = state.isAnalyzing
                    )

                    ForensicEvidenceDispatchCard(edgeVision = liveEdgeVision)

                    ReportReviewSummaryCard(
                        state = state,
                        onEditDetailsClick = { currentStep = 2 }
                    )

                    OfflineModeToggleSection(
                        isOffline = state.isOfflineMode,
                        onToggleOffline = onToggleOfflineMode
                    )

                    PrivacyNoticeCard(category = state.selectedCategory)
                }
            }
        }
    }
}

@Composable
fun WizardStepperBar(
    currentStep: Int,
    onStepClick: (Int) -> Unit
) {
    val steps = listOf(
        1 to "1. Category",
        2 to "2. Scene & Photo",
        3 to "3. AI Review"
    )
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        steps.forEach { (stepNum, title) ->
            val isActive = currentStep == stepNum
            val isCompleted = currentStep > stepNum
            val indicatorColor = when {
                isActive -> MaterialTheme.colorScheme.primary
                isCompleted -> TierSafe
                else -> MaterialTheme.colorScheme.surfaceVariant
            }
            Column(
                modifier = Modifier
                    .weight(1f)
                    .clip(RoundedCornerShape(8.dp))
                    .clickable { onStepClick(stepNum) }
                    .padding(vertical = 4.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    Box(
                        modifier = Modifier
                            .size(20.dp)
                            .background(indicatorColor, CircleShape),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = if (isCompleted) "✓" else stepNum.toString(),
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = if (isActive || isCompleted) Color.White else MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    Text(
                        text = title,
                        fontSize = 11.sp,
                        fontWeight = if (isActive) FontWeight.Bold else FontWeight.Medium,
                        color = if (isActive) MaterialTheme.colorScheme.onSurface
                        else if (isCompleted) TierSafe
                        else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(4.dp)
                        .background(indicatorColor, RoundedCornerShape(2.dp))
                )
            }
        }
    }
}

@Composable
fun SelectedCategoryHeaderBar(
    selectedCategory: ReportCategory,
    onChangeClick: () -> Unit
) {
    val (icon, label) = when (selectedCategory) {
        ReportCategory.ATTRACTANT_FRUIT_TREE -> "🍊" to "Fruit Tree / Crop Attractant"
        ReportCategory.WILDLIFE_SIGHTING -> "🐻" to "Live Wildlife Sighting"
        ReportCategory.PHYSICAL_TRACE -> "🐾" to "Physical Trace / Sign"
        ReportCategory.ROADKILL -> "🐗" to "Roadkill Carcass"
        ReportCategory.LIVESTOCK_DEPREDATION -> "🐅" to "Livestock Attack"
    }
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 10.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Text(icon, fontSize = 20.sp)
                Column {
                    Text(
                        "SELECTED CATEGORY",
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    )
                    Text(label, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                }
            }
            OutlinedButton(
                onClick = onChangeClick,
                shape = RoundedCornerShape(8.dp)
            ) {
                Text("Change", fontSize = 11.sp)
            }
        }
    }
}

@Composable
fun ReportReviewSummaryCard(
    state: IncidentReportUiState,
    onEditDetailsClick: () -> Unit
) {
    val categoryLabel = when (state.selectedCategory) {
        ReportCategory.ATTRACTANT_FRUIT_TREE -> "🍊 Fruit Tree / Crop"
        ReportCategory.WILDLIFE_SIGHTING -> "🐻 Live Sighting"
        ReportCategory.PHYSICAL_TRACE -> "🐾 Physical Trace"
        ReportCategory.ROADKILL -> "🐗 Roadkill Carcass"
        ReportCategory.LIVESTOCK_DEPREDATION -> "🐅 Livestock Attack"
    }
    val detailLabel = when (state.selectedCategory) {
        ReportCategory.WILDLIFE_SIGHTING -> state.selectedSpecies.name
        ReportCategory.PHYSICAL_TRACE -> state.selectedTraceType
        ReportCategory.ROADKILL -> "${state.selectedDecayState.name} (${state.selectedSpecies.name})"
        ReportCategory.ATTRACTANT_FRUIT_TREE ->
            if (state.isNearSchoolRoute) "${state.selectedPropertyTag.name} • School Route" else state.selectedPropertyTag.name
        ReportCategory.LIVESTOCK_DEPREDATION -> "72h Tiger Caching Buffer"
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("Report Summary", fontSize = 14.sp, fontWeight = FontWeight.Bold)
                Text(
                    text = "Edit Details",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.clickable { onEditDetailsClick() }
                )
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                        .padding(10.dp)
                ) {
                    Text("Category", fontSize = 10.sp, color = Color.Gray)
                    Text(categoryLabel, fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                        .padding(10.dp)
                ) {
                    Text("Scene Detail", fontSize = 10.sp, color = Color.Gray)
                    Text(detailLabel, fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                        .padding(10.dp)
                ) {
                    Text("GPS Coordinates", fontSize = 10.sp, color = Color.Gray)
                    Text("${state.location.latitude}, ${state.location.longitude}", fontSize = 11.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                        .padding(10.dp)
                ) {
                    Text("Evidence", fontSize = 10.sp, color = Color.Gray)
                    Text("📷 ${state.photoCount} Photo (EXIF)", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}


@Composable
fun OfflineQueueBanner(
    queueCount: Int,
    onSyncClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = TierWarning.copy(alpha = 0.15f)),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Row(
            modifier = Modifier
                .padding(14.dp)
                .fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(
                modifier = Modifier.weight(1f),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("📦", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(10.dp))
                Column {
                    Text(
                        text = "Pending Offline Queue: $queueCount report(s)",
                        fontWeight = FontWeight.Bold,
                        fontSize = 13.sp,
                        color = TierWarning
                    )
                    Text(
                        text = "Cached locally. Synchronizes automatically when back online.",
                        fontSize = 11.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                    )
                }
            }
            Spacer(modifier = Modifier.width(8.dp))
            Button(
                onClick = onSyncClick,
                colors = ButtonDefaults.buttonColors(containerColor = TierWarning),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text("Sync Now", fontSize = 11.sp, color = Color.White, fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
fun PhotoMetadataCard(
    photoCount: Int,
    location: GeoLocation,
    edgeVision: EdgeVisionAnalysisResult = EdgeVisionEngine.analyzeFrame(),
    onAnalyzeClick: () -> Unit = {},
    onRetakePhotoClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(10.dp)
                            .background(TierDanger, CircleShape)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "AI REC · 4K 60FPS · YOLOv8n INT8 (${edgeVision.inferenceLatencyMs}ms)",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                OutlinedButton(
                    onClick = onRetakePhotoClick,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text("Retake", fontSize = 11.sp)
                }
            }

            // Viewfinder Bounding Box & Optical Range Telemetry Box
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFF090D16), RoundedCornerShape(10.dp))
                    .border(1.dp, MaterialTheme.colorScheme.primary.copy(alpha = 0.5f), RoundedCornerShape(10.dp))
                    .padding(12.dp)
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "🎯 ${edgeVision.primaryDetection.taxonomy} · ${(edgeVision.primaryDetection.confidence * 1000).toInt() / 10.0}%",
                            fontSize = 13.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFF38BDF8)
                        )
                        Text(
                            text = "RAW",
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                            color = TierSafe
                        )
                    }
                    Text(
                        text = "RANGE: ${edgeVision.opticalRange.rangeMeters}m ± ${edgeVision.opticalRange.uncertaintyMeters}m · TILT: ${edgeVision.opticalRange.tiltDegrees}°",
                        fontSize = 11.sp,
                        color = Color(0xFF38BDF8),
                        fontWeight = FontWeight.SemiBold
                    )
                    Text(
                        text = "GPS: ${location.latitude}° N, ${location.longitude}° E · ALT ${edgeVision.altitudeMeters}m",
                        fontSize = 11.sp,
                        color = TierSafe
                    )
                    Text(
                        text = "PRIVACY MASK: FACES & VEHICLES BLURRED (${edgeVision.blurredRegionsCount} regions)",
                        fontSize = 10.sp,
                        color = Color.Gray
                    )
                }
            }

            Button(
                onClick = onAnalyzeClick,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(10.dp),
                colors = ButtonDefaults.buttonColors(containerColor = TierDanger)
            ) {
                Text(
                    text = "🎯 ANALYZE (YOLOv8n TFLite · ${edgeVision.modelName})",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }
        }
    }
}

@Composable
fun PredatorCachingRiskBanner(edgeVision: EdgeVisionAnalysisResult) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.5.dp, TierDanger, RoundedCornerShape(14.dp)),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color(0xFF450A0A))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(
                text = "⚠️ PREDATOR CACHING RISK DETECTED",
                fontSize = 13.sp,
                fontWeight = FontWeight.Bold,
                color = Color(0xFFFCA5A5)
            )
            Text(
                text = "Carcass covered with leaf mulch and brush (Leaf Cover: ${(edgeVision.cachingEvaluation.leafCoverScore * 100).toInt()}% > 65% threshold). High probability of defensive bear returning within 24h.",
                fontSize = 11.sp,
                color = Color.White
            )
            Text(
                text = "Immediate evacuation perimeter: ${edgeVision.cachingEvaluation.evacRadiusMeters} meters · Mode: ${edgeVision.cachingEvaluation.audioMode}",
                fontSize = 11.sp,
                fontWeight = FontWeight.Bold,
                color = Color(0xFFFCA5A5)
            )
        }
    }
}

@Composable
fun ForensicEvidenceDispatchCard(edgeVision: EdgeVisionAnalysisResult) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(
                text = "EVIDENCE ATTACHMENT & MUNICIPAL DISPATCH PAYLOAD",
                fontSize = 11.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "📷 ${edgeVision.evidenceFilename} · SHA256: ${edgeVision.sha256.take(8)}...${edgeVision.sha256.takeLast(6)} (Tamper-Proof)",
                fontSize = 11.sp,
                fontWeight = FontWeight.SemiBold
            )
            Text(
                text = "✓ Face/Vehicle Privacy Mask Applied (${edgeVision.blurredRegionsCount} regions blurred)",
                fontSize = 11.sp,
                color = TierSafe
            )
            Text(
                text = edgeVision.geoJsonPayload,
                fontSize = 10.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.75f)
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CategorySelectorSection(
    selectedCategory: ReportCategory,
    onCategoryChange: (ReportCategory) -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Text(
            text = "Step 1 • What did you observe?",
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.primary
        )
        Text(
            text = "Select the report type to pre-configure the AI forensic model and required scene details.",
            fontSize = 12.sp,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
        )

        val options = listOf(
            ReportCategory.ATTRACTANT_FRUIT_TREE to listOf("🍊", "Fruit Tree / Crop Attractant", "Unharvested persimmon (kaki), sugarcane, or orchard drawing wildlife.", "Attractant Map", "false"),
            ReportCategory.WILDLIFE_SIGHTING to listOf("🐻", "Live Wildlife Sighting", "Direct visual encounter with bear, tiger, elephant, boar, or deer.", "Encounter", "false"),
            ReportCategory.PHYSICAL_TRACE to listOf("🐾", "Physical Trace / Sign", "Footprints, claw marks on tree bark, scat/droppings, or crop damage.", "Forensic Sign", "false"),
            ReportCategory.ROADKILL to listOf("🐗", "Roadkill Carcass", "Roadside carcass acting as an olfactory predator magnet (800m lure zone).", "⚠️ Predator Lure", "true"),
            ReportCategory.LIVESTOCK_DEPREDATION to listOf("🐅", "Livestock Attack", "Domestic herd predation triggering an immediate 72-hour Tiger Caching Buffer.", "🚨 Tier 3 Danger", "true")
        )

        options.forEach { (category, meta) ->
            val icon = meta[0]
            val label = meta[1]
            val subtitle = meta[2]
            val badgeText = meta[3]
            val isAlert = meta[4] == "true"
            val isSelected = selectedCategory == category
            val highlightColor = if (isAlert) TierDanger else MaterialTheme.colorScheme.primary

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onCategoryChange(category) }
                    .border(
                        width = if (isSelected) 2.dp else 1.dp,
                        color = if (isSelected) highlightColor else MaterialTheme.colorScheme.outline.copy(alpha = 0.3f),
                        shape = RoundedCornerShape(14.dp)
                    ),
                shape = RoundedCornerShape(14.dp),
                colors = CardDefaults.cardColors(
                    containerColor = if (isSelected) highlightColor.copy(alpha = 0.12f)
                    else MaterialTheme.colorScheme.surface
                )
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(14.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalAlignment = Alignment.Top
                ) {
                    Box(
                        modifier = Modifier
                            .size(44.dp)
                            .background(
                                if (isSelected) highlightColor.copy(alpha = 0.2f)
                                else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.6f),
                                RoundedCornerShape(12.dp)
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(icon, fontSize = 22.sp)
                    }

                    Column(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(label, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                            Box(
                                modifier = Modifier
                                    .background(
                                        if (isAlert) TierDanger.copy(alpha = 0.15f)
                                        else MaterialTheme.colorScheme.surfaceVariant,
                                        RoundedCornerShape(6.dp)
                                    )
                                    .padding(horizontal = 6.dp, vertical = 2.dp)
                            ) {
                                Text(
                                    text = badgeText,
                                    fontSize = 10.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = if (isAlert) TierDanger else MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                        }
                        Text(
                            text = subtitle,
                            fontSize = 11.sp,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                            lineHeight = 16.sp
                        )
                    }

                    Box(
                        modifier = Modifier
                            .size(20.dp)
                            .background(
                                if (isSelected) highlightColor else MaterialTheme.colorScheme.surfaceVariant,
                                CircleShape
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        if (isSelected) {
                            Text("✓", fontSize = 11.sp, fontWeight = FontWeight.Bold, color = Color.White)
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun GeminiTriageCard(
    category: ReportCategory,
    result: GeminiTriageResult?,
    isAnalyzing: Boolean
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = "Instant Gemini 3.8 Flash Triage",
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "AI Forensic Verification",
                        fontSize = 11.sp,
                        color = Color.Gray
                    )
                }

                if (isAnalyzing) {
                    CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp)
                } else if (result != null) {
                    Box(
                        modifier = Modifier
                            .background(TierSafe.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 8.dp, vertical = 2.dp)
                    ) {
                        Text(
                            text = "${(result.confidence * 100).toInt()}% Match",
                            color = TierSafe,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }

            if (result != null) {
                Text(
                    text = result.detectedLabel ?: "Object Identified",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )

                result.traceType?.let { trace ->
                    Text(
                        text = "Forensic Trace: $trace",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                result.traumaType?.let { trauma ->
                    Text(
                        text = "Trauma Analysis: $trauma",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = TierDanger
                    )
                }

                result.notes?.let { notes ->
                    Text(
                        text = notes,
                        fontSize = 13.sp,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f),
                        lineHeight = 18.sp
                    )
                }

                // Forensic Tags Chips
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    result.forensicTags.forEach { tag ->
                        Box(
                            modifier = Modifier
                                .background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(8.dp))
                                .padding(horizontal = 8.dp, vertical = 4.dp)
                        ) {
                            Text(text = "#$tag", fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            } else if (!isAnalyzing) {
                Text(
                    text = "Awaiting photo analysis preview...",
                    fontSize = 13.sp,
                    color = Color.Gray
                )
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WildlifeSightingSection(
    selectedSpecies: WildlifeSpecies,
    onSpeciesChange: (WildlifeSpecies) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text(
                text = "Wildlife Species Observed",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold
            )

            val speciesList = listOf(
                WildlifeSpecies.BEAR to "Asian Black Bear",
                WildlifeSpecies.TIGER to "Sumatran Tiger",
                WildlifeSpecies.ELEPHANT to "Asian Elephant",
                WildlifeSpecies.WILD_BOAR to "Wild Boar",
                WildlifeSpecies.DEER to "Sika Deer"
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                speciesList.forEach { (sp, label) ->
                    FilterChip(
                        selected = selectedSpecies == sp,
                        onClick = { onSpeciesChange(sp) },
                        label = { Text(label, fontSize = 12.sp) },
                        shape = RoundedCornerShape(12.dp)
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PhysicalTraceContextSection(
    selectedTraceType: String,
    onTraceTypeChange: (String) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text(
                text = "Physical Trace Classification (Gemini 3.8 Forensics)",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold
            )
            Text(
                text = "Select primary physical trace observable at the scene:",
                fontSize = 12.sp,
                color = Color.Gray
            )

            val traceTypes = listOf(
                "CLAW_MARK_BARK_STRIPPING" to "Claw Marks (Tree Bark)",
                "PUGMARK_SPREAD" to "Pugmark / Footprint",
                "SCAT_VEGETATIVE_ACORN" to "Scat / Droppings",
                "HAIR_RUB_RESIDUE" to "Hair Rub / Tree Scratch",
                "CROP_TRAMPLING" to "Crop Trampling / Damage"
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                traceTypes.forEach { (typeKey, label) ->
                    FilterChip(
                        selected = selectedTraceType == typeKey,
                        onClick = { onTraceTypeChange(typeKey) },
                        label = { Text(label, fontSize = 11.sp) },
                        shape = RoundedCornerShape(12.dp)
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RoadkillContextSection(
    selectedSpecies: WildlifeSpecies,
    selectedDecayState: DecayState,
    onSpeciesChange: (WildlifeSpecies) -> Unit,
    onDecayStateChange: (DecayState) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text(
                text = "Roadkill Carcass Mitigation Logging (Issue #3)",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                color = TierDanger
            )

            Text(
                text = "Carcass Species:",
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                listOf(
                    WildlifeSpecies.WILD_BOAR to "Wild Boar",
                    WildlifeSpecies.DEER to "Deer",
                    WildlifeSpecies.UNKNOWN to "Other Small Mammal"
                ).forEach { (sp, label) ->
                    FilterChip(
                        selected = selectedSpecies == sp,
                        onClick = { onSpeciesChange(sp) },
                        label = { Text(label, fontSize = 11.sp) },
                        shape = RoundedCornerShape(12.dp)
                    )
                }
            }

            Text(
                text = "Decay State (Olfactory Lure Index):",
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                DecayState.values().forEach { decay ->
                    val (label, threat) = when (decay) {
                        DecayState.FRESH -> "Fresh (0-6h)" to "Threat: 0.60"
                        DecayState.BLOATED -> "Bloated (High Lure)" to "Threat: 0.88"
                        DecayState.DECOMPOSING -> "Decomposing" to "Threat: 0.75"
                        DecayState.SKELETAL -> "Skeletal" to "Threat: 0.20"
                    }
                    FilterChip(
                        selected = selectedDecayState == decay,
                        onClick = { onDecayStateChange(decay) },
                        label = { Text("$label ($threat)", fontSize = 11.sp) },
                        shape = RoundedCornerShape(12.dp)
                    )
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(TierDanger.copy(alpha = 0.10f), RoundedCornerShape(8.dp))
                    .padding(10.dp)
            ) {
                Text(
                    text = "⚠️ Registering carcass triggers municipal sanitation dispatch and activates an 800m predator lure geofence.",
                    fontSize = 11.sp,
                    color = TierDanger,
                    fontWeight = FontWeight.Medium
                )
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FruitTreeContextSection(
    selectedTag: PropertyContextTag,
    isNearSchoolRoute: Boolean,
    onPropertyTagChange: (PropertyContextTag) -> Unit,
    onNearSchoolRouteChange: (Boolean) -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = "Property & Hazard Context (Kaki Map)",
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold
        )

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            PropertyContextTag.values().forEach { tag ->
                val label = when (tag) {
                    PropertyContextTag.VACANT_HOUSE -> "Vacant House (Akiya)"
                    PropertyContextTag.ROADSIDE_UNMANAGED -> "Roadside Unmanaged"
                    PropertyContextTag.ACTIVE_FARM_BORDER -> "Farm Border"
                    PropertyContextTag.PUBLIC_LAND -> "Public Land"
                    PropertyContextTag.UNKNOWN -> "Other"
                }

                FilterChip(
                    selected = selectedTag == tag,
                    onClick = { onPropertyTagChange(tag) },
                    label = { Text(label, fontSize = 12.sp) },
                    shape = RoundedCornerShape(16.dp)
                )
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(10.dp))
                .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
                .clickable { onNearSchoolRouteChange(!isNearSchoolRoute) }
                .padding(12.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "Located along School Route / Residential Area",
                fontSize = 13.sp,
                fontWeight = FontWeight.Medium
            )
            Box(
                modifier = Modifier
                    .size(20.dp)
                    .background(
                        if (isNearSchoolRoute) TierDanger else Color.LightGray,
                        CircleShape
                    ),
                contentAlignment = Alignment.Center
            ) {
                if (isNearSchoolRoute) {
                    Text("✓", color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LivestockDepredationSection(
    selectedSpecies: WildlifeSpecies,
    onSpeciesChange: (WildlifeSpecies) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text(
                text = "Livestock Depredation Report (Sumatran Tiger Corridor)",
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                color = TierDanger
            )
            Text(
                text = "Target Livestock Attacked:",
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold
            )

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                listOf(
                    WildlifeSpecies.TIGER to "Cattle (Bovine)",
                    WildlifeSpecies.BEAR to "Goat / Sheep"
                ).forEach { (sp, label) ->
                    FilterChip(
                        selected = selectedSpecies == sp,
                        onClick = { onSpeciesChange(sp) },
                        label = { Text(label, fontSize = 11.sp) },
                        shape = RoundedCornerShape(12.dp)
                    )
                }
            }
        }
    }
}

@Composable
fun OfflineModeToggleSection(
    isOffline: Boolean,
    onToggleOffline: (Boolean) -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(10.dp))
            .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f))
            .clickable { onToggleOffline(!isOffline) }
            .padding(12.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column(modifier = Modifier.weight(1f)) {
            Text("Simulate Offline Mode (Field Worker In-Hinterland)", fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
            Text("Reports queue locally in StateFlow/SQLite and sync upon reconnecting", fontSize = 11.sp, color = Color.Gray)
        }
        Spacer(modifier = Modifier.width(8.dp))
        Box(
            modifier = Modifier
                .size(22.dp)
                .background(if (isOffline) TierWarning else Color.LightGray, CircleShape),
            contentAlignment = Alignment.Center
        ) {
            if (isOffline) {
                Text("✓", color = Color.White, fontSize = 12.sp, fontWeight = FontWeight.Bold)
            }
        }
    }
}

@Composable
fun PrivacyNoticeCard(category: ReportCategory) {
    val message = if (category == ReportCategory.ATTRACTANT_FRUIT_TREE) {
        "Private Property Protection: Exact coordinates are shared strictly with the official Municipal Wildlife Control Team (Chōjū Higai Taisaku Jisshitai) for legal dispatch. Public maps display generalized hazard heatmaps."
    } else {
        "Data Verification: Reports are encrypted and synchronized with Cloud Firestore. Gemini 3.8 Flash AI verifies forensic signatures to eliminate spam."
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(10.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f))
    ) {
        Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Text("ℹ️", fontSize = 14.sp)
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = message,
                fontSize = 11.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                lineHeight = 16.sp
            )
        }
    }
}
