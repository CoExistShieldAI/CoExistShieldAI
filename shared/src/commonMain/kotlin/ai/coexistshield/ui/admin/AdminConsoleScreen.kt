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
import androidx.compose.material3.LinearProgressIndicator
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
    val distanceToSchoolMeters: Int = 120,
    val ownerConsentStatus: String = "GRANTED",
    val permitId: String = "PERMIT-MRK-2026-1120",
    val clearanceOrder: Int = 1,
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
    // Japan Municipal GCF Data
    val totalUnmanagedTrees: Int = 840,
    val top5PercentPriorityTrees: List<PriorityKakiItem> = listOf(
        PriorityKakiItem("kt_1120", "盛岡市高松3丁目12-4 • 空き家", 0.95, true, 120, "GRANTED", "PERMIT-MRK-2026-1120", 1, "ABUNDANT (Falling)"),
        PriorityKakiItem("kt_0842", "盛岡市上米内字中野 • 林縁放置果樹園", 0.91, true, 240, "ORDINANCE_CLEARANCE", "PERMIT-MRK-2026-0842", 2, "ABUNDANT"),
        PriorityKakiItem("kt_0319", "盛岡市浅岸字二ツ森 • 住宅街裏山", 0.88, true, 310, "PENDING_CONSENT", "PERMIT-MRK-2026-0319", 3, "MODERATE")
    ),
    val gcfFundingGoalYen: Long = 15000000L,
    val gcfCurrentRaisedYen: Long = 11850000L,
    val gcfFundingProgressPct: Double = 79.0,
    val gcfDonorCount: Int = 482,
    val gcfTreesClearedCount: Int = 214,
    val gcfTargetTreesCount: Int = 420,
    val gcfSchoolChildrenProtected: Int = 3420,
    val gcfAttractantReductionPct: Double = 64.8,
    // APAC Plantation Data
    val pendingCarcasses: List<CarcassDispatchItem> = listOf(
        CarcassDispatchItem("rk_4412", "Sumatra Riau • Highway 1 South", "Wild Boar", "Bloated", 0.88),
        CarcassDispatchItem("ls_sumatra_01", "Desa Bukit Kesuma • Pasture A", "Cattle Kill-Site (Tiger 72h)", "Active Return Risk", 0.95)
    ),
    val rspoZeroFatalityDays: Int = 184,
    val rspoHcvMonitoredHectares: Double = 42500.0,
    val rspoSafePassagesCount: Int = 342,
    val rspoAuditSignature: String = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    // System AI Health & Multi-Source Corroboration Data
    val lastRetrainTimestamp: String = "2026-10-14 02:30 UTC",
    val modelAucScore: Double = 0.86,
    val activeGeofenceCount: Int = 18,
    val isRetrainingInProgress: Boolean = false,
    val geminiModelVersion: String = "gemini-3.8-flash",
    val activeCausalStreamsCount: Int = 5,
    val causalCorroborationConfidence: Double = 0.978,
    val causalCorroboratedSources: List<String> = listOf("AGRICULTURAL_HARVEST", "MUNICIPAL_ATTRACTANT", "SATELLITE_THERMAL"),
    val causalAiJustification: String = "TIER_2_WARNING: Corroborated unharvested persimmon orchard and morning waste schedule (combined confidence: 97.8%) indicating imminent bear intrusion risk.",
    val nlpExtractedSpecies: String = "Ursus thibetanus (Asian Black Bear)",
    val nlpExtractedUrgency: String = "HIGH",
    val nlpExtractedLandmark: String = "盛岡市松ヶ枝町付近 (小学校通学路南300m)",
    // BigQuery GIS Real-Time Streaming Data Lake (Issue #20)
    val bigQuerySyncStatus: String = "ACTIVE_STREAMING",
    val bigQueryReplicationLagSeconds: Double = 12.4,
    val bigQuerySlaTargetSeconds: Int = 60,
    val bigQueryStagedSamplesCount: Int = 1420,
    val bigQuerySpatialIndexing: String = "Uber H3 Res 8 (~460m edge) & S2 Level 13"
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
    // 1. GCF Progress & Budget Metrics Card (Furusato Chosei Integration)
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.45f))
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "ふるさと納税 GCF 連携: 放置柿緊急伐採プロジェクト",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.primary
                )
                Box(
                    modifier = Modifier
                        .background(TierSafe.copy(alpha = 0.2f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 2.dp)
                ) {
                    Text("盛岡市 (03201)", color = TierSafe, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                }
            }

            // Financial Goal & Raised Progress
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "調達資金額: ¥${state.gcfCurrentRaisedYen / 10000}万 / ¥${state.gcfFundingGoalYen / 10000}万円",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Text(
                        text = "${state.gcfFundingProgressPct}% 達成 (${state.gcfDonorCount}名)",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Black,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                LinearProgressIndicator(
                    progress = { (state.gcfFundingProgressPct / 100.0).toFloat().coerceIn(0f, 1f) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(8.dp)
                        .clip(RoundedCornerShape(4.dp)),
                    color = MaterialTheme.colorScheme.primary
                )
            }

            // Key Quantified Impact Metrics
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("伐採実績", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.gcfTreesClearedCount} / ${state.gcfTargetTreesCount}本", fontSize = 16.sp, fontWeight = FontWeight.Black)
                    Text("残 ${state.gcfTargetTreesCount - state.gcfTreesClearedCount}本", fontSize = 10.sp, color = Color.Gray)
                }
                Column {
                    Text("遭遇リスク削減", fontSize = 11.sp, color = Color.Gray)
                    Text("-${state.gcfAttractantReductionPct}%", fontSize = 16.sp, fontWeight = FontWeight.Black, color = TierSafe)
                    Text("空間統計再計算済", fontSize = 10.sp, color = TierSafe)
                }
                Column {
                    Text("保護児童数", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.gcfSchoolChildrenProtected}名", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                    Text("通学路沿線", fontSize = 10.sp, color = Color.Gray)
                }
                Column {
                    Text("単位伐採コスト", fontSize = 11.sp, color = Color.Gray)
                    Text("¥35,000", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                    Text("全額GCF充当", fontSize = 10.sp, color = Color.Gray)
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
            Text("鳥獣被害対策実施隊 伐採優先度リスト (Top 5% Priority)", fontSize = 15.sp, fontWeight = FontWeight.Bold)
            Text("通学路距離・クマ誘引半径・公的伐採許可書 (Permit ID) 連携", fontSize = 11.sp, color = Color.Gray)
        }
        Box(
            modifier = Modifier
                .background(TierDanger.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                .padding(horizontal = 8.dp, vertical = 4.dp)
        ) {
            Text("優先度順位付", color = TierDanger, fontSize = 11.sp, fontWeight = FontWeight.Bold)
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
                val isFelled = tree.status == TreeManagementStatus.FELLED || tree.status == TreeManagementStatus.PRUNED

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Box(
                            modifier = Modifier
                                .background(MaterialTheme.colorScheme.primary, RoundedCornerShape(4.dp))
                                .padding(horizontal = 6.dp, vertical = 2.dp)
                        ) {
                            Text("#${tree.clearanceOrder}", color = Color.White, fontSize = 11.sp, fontWeight = FontWeight.Bold)
                        }
                        Text(tree.addressHint, fontWeight = FontWeight.Bold, fontSize = 14.sp)
                    }
                    Box(
                        modifier = Modifier
                            .background(if (isFelled) TierSafe else TierDanger, RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 2.dp)
                    ) {
                        Text(
                            if (isFelled) "伐採済 0%" else "Hazard ${(tree.hazardScore * 100).toInt()}%",
                            color = Color.White,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    if (tree.isNearSchool) {
                        Text("⚠️ 通学路から${tree.distanceToSchoolMeters}m", fontSize = 11.sp, color = TierWarning, fontWeight = FontWeight.SemiBold)
                    }
                    Text("• 実り: ${tree.fruitLoad}", fontSize = 11.sp, color = Color.Gray)
                }

                // Permit Status Badge
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    val (consentBadgeColor, consentText) = when (tree.ownerConsentStatus) {
                        "GRANTED" -> TierSafe to "承諾済: ${tree.permitId}"
                        "ORDINANCE_CLEARANCE" -> TierWarning to "条例代執行: ${tree.permitId}"
                        else -> Color.Gray to "所有者承諾手続き中"
                    }
                    Box(
                        modifier = Modifier
                            .background(consentBadgeColor.copy(alpha = 0.15f), RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 2.dp)
                    ) {
                        Text(consentText, color = consentBadgeColor, fontSize = 10.sp, fontWeight = FontWeight.Bold)
                    }
                }

                Spacer(modifier = Modifier.height(4.dp))

                // Action Row with Role-Based Access Control (RBAC)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End
                ) {
                    val isAuthorizedOfficer = state.userRole == UserRole.OFFICIAL || state.userRole == UserRole.ADMIN

                    if (isFelled) {
                        OutlinedButton(
                            onClick = {},
                            enabled = false,
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text("✅ 伐採執行済 (リスク0%)", fontSize = 12.sp, color = TierSafe)
                        }
                    } else if (isAuthorizedOfficer) {
                        Button(
                            onClick = { onMarkTreePruned(tree.id) },
                            shape = RoundedCornerShape(8.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = TierSafe)
                        ) {
                            Text("公的伐採・剪定を執行 (官民連携)", fontSize = 12.sp, fontWeight = FontWeight.Bold)
                        }
                    } else {
                        OutlinedButton(
                            onClick = {},
                            enabled = false,
                            shape = RoundedCornerShape(8.dp)
                        ) {
                            Text("公的伐採隊専用 (無断伐採防止)", fontSize = 11.sp, color = Color.Gray)
                        }
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

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("Monitored HCV Area", fontSize = 11.sp, color = Color.Gray)
                    Text("${(state.rspoHcvMonitoredHectares).toInt()} ha", fontSize = 14.sp, fontWeight = FontWeight.Bold)
                }
                Column {
                    Text("Safe Corridor Passages", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.rspoSafePassagesCount} Herds", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
                Column {
                    Text("Compliance Score", fontSize = 11.sp, color = Color.Gray)
                    Text("100% Zero-Harm", fontSize = 14.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
            }

            Text(
                text = "Audit Signature: ${state.rspoAuditSignature.take(28)}... (Tamper-Proof SHA-256)",
                fontSize = 10.sp,
                color = Color.Gray
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

    // Card 1b: BigQuery GIS Real-Time Streaming Data Lake (Issue #20)
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
                Text("BigQuery GIS Streaming Data Lake", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                Box(
                    modifier = Modifier
                        .background(TierSafe.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 2.dp)
                ) {
                    Text(state.bigQuerySyncStatus, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
            }

            Text(
                text = "Real-time event replication from Cloud Firestore to BigQuery GIS with automated spatio-temporal feature enrichment.",
                fontSize = 12.sp,
                color = Color.Gray
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("Replication Lag", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.bigQueryReplicationLagSeconds}s (< ${state.bigQuerySlaTargetSeconds}s SLA)", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
                Column {
                    Text("Staged Samples", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.bigQueryStagedSamplesCount} Samples", fontSize = 13.sp, fontWeight = FontWeight.Black)
                }
                Column {
                    Text("Spatial Grid", fontSize = 11.sp, color = Color.Gray)
                    Text("H3 L8 / S2 L13", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.primary)
                }
            }
        }
    }

    // Card 2: Multi-Source Causal Corroboration Engine (Gemini 3.8 Flash & 5 Environmental Streams)
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
                Text("Multi-Source Causal Corroboration Engine", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                Box(
                    modifier = Modifier
                        .background(MaterialTheme.colorScheme.primaryContainer, RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 2.dp)
                ) {
                    Text(state.geminiModelVersion, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.primary)
                }
            }

            Text(
                text = "Autonomous Bayesian fusion across 5 environmental streams: Agricultural, Municipal Waste, Satellite FIRMS, Social NLP, and Traffic.",
                fontSize = 12.sp,
                color = Color.Gray
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text("Active Feeds", fontSize = 11.sp, color = Color.Gray)
                    Text("${state.activeCausalStreamsCount} / 5 Feeds", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = TierSafe)
                }
                Column {
                    Text("Sensor Fusion", fontSize = 11.sp, color = Color.Gray)
                    Text("${(state.causalCorroborationConfidence * 100).toInt()}% Conf", fontSize = 13.sp, fontWeight = FontWeight.Black, color = TierWarning)
                }
                Column {
                    Text("Triggered Tier", fontSize = 11.sp, color = Color.Gray)
                    Text("TIER 2 WARNING", fontSize = 13.sp, fontWeight = FontWeight.Bold, color = TierWarning)
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                    .padding(10.dp)
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Text("Gemini 3.8 Flash Causal Justification:", fontSize = 11.sp, fontWeight = FontWeight.Bold)
                    Text(state.causalAiJustification, fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }

            // Real-time NLP Sighting Parser Sample
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(TierNotice.copy(alpha = 0.1f), RoundedCornerShape(8.dp))
                    .padding(10.dp)
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text("Live NLP Entity Extraction (Municipal Radio)", fontSize = 11.sp, fontWeight = FontWeight.Bold, color = TierNotice)
                        Text(state.nlpExtractedUrgency, fontSize = 11.sp, fontWeight = FontWeight.Bold, color = TierDanger)
                    }
                    Text(
                        text = "Species: ${state.nlpExtractedSpecies} • Loc: ${state.nlpExtractedLandmark}",
                        fontSize = 11.sp,
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }
            }
        }
    }
}
