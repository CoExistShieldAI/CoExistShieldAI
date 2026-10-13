package ai.coexistshield.ui.admin

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.FruitLoadLevel
import ai.coexistshield.model.KakiTree
import ai.coexistshield.model.PropertyContextTag
import ai.coexistshield.model.TreeManagementStatus
import ai.coexistshield.model.UserRole
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
import androidx.compose.material3.Divider
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
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

enum class AdminTab(val title: String) {
    JAPAN_MUNICIPAL("Japan: Kaki & GCF"),
    APAC_PLANTATION("APAC: Plantations & RSPO"),
    SYSTEM_AI("AI System Health")
}

data class PriorityKakiItem(
    val id: String,
    val addressHint: String,
    val hazardScore: Double,
    val isNearSchool: Boolean,
    val fruitLoad: String,
    val status: TreeManagementStatus = TreeManagementStatus.UNMANAGED
)

data class CarcassDispatchItem(
    val id: String,
    val locationHint: String,
    val species: String,
    val decayState: String,
    val threatScore: Double,
    val isDispatched: Boolean = false
)

data class AdminConsoleUiState(
    val activeTab: AdminTab = AdminTab.JAPAN_MUNICIPAL,
    val userRole: UserRole = UserRole.OFFICIAL,
    // Japan Municipal Data
    val totalUnmanagedTrees: Int = 840,
    val top5PercentPriorityTrees: List<PriorityKakiItem> = listOf(
        PriorityKakiItem("kt_1120", "Iwate Morioka #12 • Vacant House", 0.95, true, "ABUNDANT (Falling)"),
        PriorityKakiItem("kt_1121", "Akita Kazuno #04 • Roadside Forest Border", 0.92, true, "ABUNDANT"),
        PriorityKakiItem("kt_1122", "Hokkaido Sapporo #8 • Aging Orchard", 0.89, false, "MODERATE")
    ),
    val gcfTreesClearedCount: Int = 42,
    val gcfAttractantReductionPct: Int = 38,
    // APAC Plantation Data
    val pendingCarcasses: List<CarcassDispatchItem> = listOf(
        CarcassDispatchItem("rk_4412", "Sumatra Riau • Highway 1 South", "Wild Boar", "Bloated", 0.88),
        CarcassDispatchItem("ls_901", "Kui Buri Forest Border • Pasture B", "Cattle Kill-Site (Tiger 72h)", "Fresh Kill", 0.94)
    ),
    val rspoZeroFatalityDays: Int = 184,
    // System AI Health Data
    val lastRetrainTimestamp: String = "2026-10-13 18:30 UTC",
    val modelAucScore: Double = 0.86,
    val activeGeofenceCount: Int = 18,
    val isRetrainingInProgress: Boolean = false
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AdminConsoleScreen(
    state: AdminConsoleUiState,
    onBackClick: () -> Unit = {},
    onTabSelect: (AdminTab) -> Unit = {},
    onMarkTreePruned: (String) -> Unit = {},
    onDispatchCarcassTruck: (String) -> Unit = {},
    onExportRspoReport: () -> Unit = {},
    onTriggerAiRetraining: () -> Unit = {}
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("Decision-Support Console", fontSize = 18.sp, fontWeight = FontWeight.Bold)
                        Spacer(modifier = Modifier.width(8.dp))
                        Box(
                            modifier = Modifier
                                .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.15f), RoundedCornerShape(8.dp))
                                .padding(horizontal = 8.dp, vertical = 2.dp)
                        ) {
                            Text(state.userRole.name, color = MaterialTheme.colorScheme.primary, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                        }
                    }
                },
                navigationIcon = {
                    Box(modifier = Modifier.padding(start = 12.dp)) {
                        OutlinedButton(
                            onClick = onBackClick,
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text("HUD", fontSize = 12.sp)
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
        ) {
            // Tab Selector Row
            TabRow(
                selectedTabIndex = state.activeTab.ordinal,
                containerColor = MaterialTheme.colorScheme.surface
            ) {
                AdminTab.values().forEach { tab ->
                    // Restrict System Tab to ADMIN only
                    if (tab != AdminTab.SYSTEM_AI || state.userRole == UserRole.ADMIN) {
                        Tab(
                            selected = state.activeTab == tab,
                            onClick = { onTabSelect(tab) },
                            text = { Text(tab.title, fontSize = 12.sp, fontWeight = FontWeight.SemiBold) }
                        )
                    }
                }
            }

            // Tab Content Body
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp)
                    .verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                when (state.activeTab) {
                    AdminTab.JAPAN_MUNICIPAL -> JapanMunicipalTab(
                        state = state,
                        onMarkTreePruned = onMarkTreePruned
                    )
                    AdminTab.APAC_PLANTATION -> ApacPlantationTab(
                        state = state,
                        onDispatchCarcassTruck = onDispatchCarcassTruck,
                        onExportRspoReport = onExportRspoReport
                    )
                    AdminTab.SYSTEM_AI -> SystemAiHealthTab(
                        state = state,
                        onTriggerAiRetraining = onTriggerAiRetraining
                    )
                }
            }
        }
    }
}

// -------------------------------------------------------------
// TAB 1: Japan Municipal Wildlife Control & GCF
// -------------------------------------------------------------
@Composable
fun JapanMunicipalTab(
    state: AdminConsoleUiState,
    onMarkTreePruned: (String) -> Unit
) {
    // 1. GCF Progress & Budget Metrics Card
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.4f))
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Text(
                text = "Government Crowdfunding (GCF) Progress",
                fontSize = 13.sp,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("Trees Cleared via GCF", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.gcfTreesClearedCount} Trees", fontSize = 18.sp, fontWeight = FontWeight.Black)
                }
                Column {
                    Text("Attractant Reduction", fontSize = 11.sp, color = Color.Gray)
                    Text("-${state.gcfAttractantReductionPct}%", fontSize = 18.sp, fontWeight = FontWeight.Black, color = TierSafe)
                }
                Column {
                    Text("Cataloged Trees", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.totalUnmanagedTrees}", fontSize = 18.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }

    // 2. Top 5% Lethal Hazard Persimmon Roster Header
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column {
            Text("Top 5% Lethal Hazard Persimmon Trees", fontSize = 15.sp, fontWeight = FontWeight.Bold)
            Text("Filtered by proximity to school routes & residential density", fontSize = 11.sp, color = Color.Gray)
        }
        Box(
            modifier = Modifier
                .background(TierDanger.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                .padding(horizontal = 8.dp, vertical = 4.dp)
        ) {
            Text("High Priority", color = TierDanger, fontSize = 11.sp, fontWeight = FontWeight.Bold)
        }
    }

    // Priority Trees List
    state.top5PercentPriorityTrees.forEach { tree ->
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            border = CardDefaults.outlinedCardBorder()
        ) {
            Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(tree.addressHint, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                    Box(
                        modifier = Modifier
                            .background(TierDanger, RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 2.dp)
                    ) {
                        Text("Hazard ${(tree.hazardScore * 100).toInt()}%", color = Color.White, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    }
                }

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    if (tree.isNearSchool) {
                        Text("⚠️ Near School Route", fontSize = 11.sp, color = TierWarning, fontWeight = FontWeight.SemiBold)
                    }
                    Text("• Fruit: ${tree.fruitLoad}", fontSize = 11.sp, color = Color.Gray)
                }

                Spacer(modifier = Modifier.height(4.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End
                ) {
                    Button(
                        onClick = { onMarkTreePruned(tree.id) },
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = TierSafe)
                    ) {
                        Text("Mark Pruned / Felled", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

// -------------------------------------------------------------
// TAB 2: APAC Plantation Operations & RSPO Compliance
// -------------------------------------------------------------
@Composable
fun ApacPlantationTab(
    state: AdminConsoleUiState,
    onDispatchCarcassTruck: (String) -> Unit,
    onExportRspoReport: () -> Unit
) {
    // 1. RSPO ESG Compliance Banner Card
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
                Text(
                    text = "RSPO & FSC ESG Wildlife Compliance",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Box(
                    modifier = Modifier
                        .background(TierSafe.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 2.dp)
                ) {
                    Text("Audit Certified", color = TierSafe, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                }
            }

            Text(
                text = "${state.rspoZeroFatalityDays} Consecutive Days — Zero Human Fatalities across Concession",
                fontSize = 14.sp,
                fontWeight = FontWeight.Black
            )

            Button(
                onClick = onExportRspoReport,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(8.dp)
            ) {
                Text("Export 1-Click RSPO Audit Package (PDF / CSV)", fontSize = 13.sp, fontWeight = FontWeight.Bold)
            }
        }
    }

    // 2. Predatory Attractant & Roadkill Clearance Queue
    Text("Active Carcass & Kill-Site Dispatch Queue", fontSize = 15.sp, fontWeight = FontWeight.Bold)
    Text("Decomposing carcasses act as predator magnets. Prompt removal prevents attacks.", fontSize = 11.sp, color = Color.Gray)

    state.pendingCarcasses.forEach { item ->
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            border = CardDefaults.outlinedCardBorder()
        ) {
            Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("${item.species} • ${item.locationHint}", fontWeight = FontWeight.Bold, fontSize = 14.sp)
                    Box(
                        modifier = Modifier
                            .background(TierDanger.copy(alpha = 0.15f), RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 2.dp)
                    ) {
                        Text("Threat ${(item.threatScore * 100).toInt()}%", color = TierDanger, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    }
                }

                Text("State: ${item.decayState}", fontSize = 12.sp, color = Color.Gray)

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End
                ) {
                    Button(
                        onClick = { onDispatchCarcassTruck(item.id) },
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = TierWarning)
                    ) {
                        Text("Dispatch Sanitation / Rangers", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}

// -------------------------------------------------------------
// TAB 3: System Health & AI Governance (Admin Only)
// -------------------------------------------------------------
@Composable
fun SystemAiHealthTab(
    state: AdminConsoleUiState,
    onTriggerAiRetraining: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        border = CardDefaults.outlinedCardBorder()
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("Autonomous Retraining Engine", fontSize = 15.sp, fontWeight = FontWeight.Bold)

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("Last Train Run", fontSize = 11.sp, color = Color.Gray)
                    Text(state.lastRetrainTimestamp, fontSize = 13.sp, fontWeight = FontWeight.SemiBold)
                }
                Column {
                    Text("Model Performance", fontSize = 11.sp, color = Color.Gray)
                    Text("AUC ${state.modelAucScore}", fontSize = 13.sp, fontWeight = FontWeight.Black, color = TierSafe)
                }
                Column {
                    Text("Active Polygons", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.activeGeofenceCount} Zones", fontSize = 13.sp, fontWeight = FontWeight.Bold)
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            Button(
                onClick = onTriggerAiRetraining,
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(8.dp),
                enabled = !state.isRetrainingInProgress
            ) {
                if (state.isRetrainingInProgress) {
                    CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp, color = Color.White)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Retraining in Progress on Cloud Run...", fontSize = 13.sp)
                } else {
                    Text("Force Retrain Models on Accumulated Incidents", fontSize = 13.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}
