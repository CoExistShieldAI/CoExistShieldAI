package ai.coexistshield.ui.report

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
    val geminiTriageResult: GeminiTriageResult? = GeminiTriageResult(
        detectedLabel = "Diospyros kaki (Persimmon)",
        commonName = "Abandoned Persimmon Tree",
        confidence = 0.96,
        forensicTags = listOf("abundant_fruit", "overripe_falling", "bear_attractant"),
        fruitLoadLevel = "ABUNDANT",
        notes = "High-calorie bear attractant cluster detected within 500m of residential structure."
    ),
    val selectedPropertyTag: PropertyContextTag = PropertyContextTag.VACANT_HOUSE,
    val isNearSchoolRoute: Boolean = true,
    val isAnalyzing: Boolean = false,
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
    onPropertyTagChange: (PropertyContextTag) -> Unit = {},
    onNearSchoolRouteChange: (Boolean) -> Unit = {},
    onRetakePhotoClick: () -> Unit = {},
    onSubmitReportClick: () -> Unit = {}
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("Forensic Incident & Tree Report", fontSize = 18.sp, fontWeight = FontWeight.Bold)
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
            // 1. Photo Viewfinder & GPS Metadata Card
            PhotoMetadataCard(
                photoCount = state.photoCount,
                location = state.location,
                onRetakePhotoClick = onRetakePhotoClick
            )

            // 2. Report Category Selector (5 distinct categories)
            CategorySelectorSection(
                selectedCategory = state.selectedCategory,
                onCategoryChange = onCategoryChange
            )

            // 3. Instant Gemini Multimodal Triage Card
            GeminiTriageCard(
                category = state.selectedCategory,
                result = state.geminiTriageResult,
                isAnalyzing = state.isAnalyzing
            )

            // 4. Context Tags Section (e.g. Vacant House, Near School Route)
            if (state.selectedCategory == ReportCategory.ATTRACTANT_FRUIT_TREE) {
                FruitTreeContextSection(
                    selectedTag = state.selectedPropertyTag,
                    isNearSchoolRoute = state.isNearSchoolRoute,
                    onPropertyTagChange = onPropertyTagChange,
                    onNearSchoolRouteChange = onNearSchoolRouteChange
                )
            }

            // 5. Privacy Notice
            PrivacyNoticeCard(category = state.selectedCategory)

            Spacer(modifier = Modifier.height(8.dp))

            // 6. Submit Button
            Button(
                onClick = onSubmitReportClick,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(52.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (state.selectedCategory == ReportCategory.LIVESTOCK_DEPREDATION || state.selectedCategory == ReportCategory.ROADKILL)
                        TierDanger else MaterialTheme.colorScheme.primary
                )
            ) {
                Text(
                    text = if (state.isQueuedOffline) "Save to Offline Queue" else "Submit Verified Report",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold
                )
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
                        text = "GPS Coordinate Locked",
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
                        text = "Instant Gemini Triage",
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "2s AI Verification",
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

@Composable
fun PrivacyNoticeCard(category: ReportCategory) {
    val message = if (category == ReportCategory.ATTRACTANT_FRUIT_TREE) {
        "Private Property Protection: Exact coordinates are shared strictly with the official Municipal Wildlife Control Team (Chōjū Higai Taisaku Jisshitai) for legal dispatch. Public maps display generalized hazard heatmaps."
    } else {
        "Data Verification: Reports are encrypted and synchronized with Cloud Firestore. Gemini AI verifies forensic signatures to eliminate spam."
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
