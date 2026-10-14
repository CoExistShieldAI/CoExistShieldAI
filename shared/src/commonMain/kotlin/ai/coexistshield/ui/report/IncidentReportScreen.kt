package ai.coexistshield.ui.report

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
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("Forensic Incident & Field Report", fontSize = 18.sp, fontWeight = FontWeight.Bold)
                },
                navigationIcon = {
                    Box(modifier = Modifier.padding(start = 12.dp)) {
                        OutlinedButton(
                            onClick = onBackClick,
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text("Cancel", fontSize = 12.sp)
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
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
            // 1. Offline Pending Queue Banner (Issue #2: Offline persistence badge)
            if (state.offlineQueueCount > 0) {
                OfflineQueueBanner(
                    queueCount = state.offlineQueueCount,
                    onSyncClick = onSyncOfflineQueueClick
                )
            }

            // 2. Photo Viewfinder & GPS Metadata Card
            PhotoMetadataCard(
                photoCount = state.photoCount,
                location = state.location,
                onRetakePhotoClick = onRetakePhotoClick
            )

            // 3. Report Category Selector (5 distinct categories)
            CategorySelectorSection(
                selectedCategory = state.selectedCategory,
                onCategoryChange = onCategoryChange
            )

            // 4. Instant Gemini Multimodal Triage Card
            GeminiTriageCard(
                category = state.selectedCategory,
                result = state.geminiTriageResult,
                isAnalyzing = state.isAnalyzing
            )

            // 5. Dynamic Context Options by Category
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

            // 6. Offline Mode Simulator Toggle
            OfflineModeToggleSection(
                isOffline = state.isOfflineMode,
                onToggleOffline = onToggleOfflineMode
            )

            // 7. Privacy & Security Notice
            PrivacyNoticeCard(category = state.selectedCategory)

            Spacer(modifier = Modifier.height(8.dp))

            // 8. Submit Button
            Button(
                onClick = onSubmitReportClick,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (state.isOfflineMode) TierWarning
                    else if (state.selectedCategory == ReportCategory.LIVESTOCK_DEPREDATION || state.selectedCategory == ReportCategory.ROADKILL)
                        TierDanger else MaterialTheme.colorScheme.primary
                )
            ) {
                Text(
                    text = if (state.isOfflineMode) "Save to Offline Queue (Local Sync)"
                    else "Submit Verified Report (Gemini AI)",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
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
    onRetakePhotoClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f))
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(10.dp)
                            .background(TierSafe, CircleShape)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "GPS Coordinates Locked (EXIF)",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = TierSafe
                    )
                }
                OutlinedButton(
                    onClick = onRetakePhotoClick,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text("Retake Photo", fontSize = 11.sp)
                }
            }

            Text(
                text = "Lat: ${location.latitude}, Long: ${location.longitude}",
                fontSize = 12.sp,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
            )
            Text(
                text = "Evidence Photos Attached: $photoCount photo(s)",
                fontSize = 12.sp,
                fontWeight = FontWeight.Medium
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
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = "Report Category",
            fontSize = 14.sp,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.onSurface
        )

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            ReportCategory.values().forEach { category ->
                val (label, isAlert) = when (category) {
                    ReportCategory.WILDLIFE_SIGHTING -> "Live Sighting" to false
                    ReportCategory.PHYSICAL_TRACE -> "Physical Trace" to false
                    ReportCategory.ROADKILL -> "Roadkill Carcass" to true
                    ReportCategory.LIVESTOCK_DEPREDATION -> "Livestock Attack" to true
                    ReportCategory.ATTRACTANT_FRUIT_TREE -> "Fruit Tree / Crop" to false
                }

                FilterChip(
                    selected = selectedCategory == category,
                    onClick = { onCategoryChange(category) },
                    label = { Text(label, fontSize = 12.sp) },
                    shape = RoundedCornerShape(16.dp),
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = if (isAlert) TierDanger.copy(alpha = 0.2f) else MaterialTheme.colorScheme.primaryContainer,
                        selectedLabelColor = if (isAlert) TierDanger else MaterialTheme.colorScheme.onPrimaryContainer
                    )
                )
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
