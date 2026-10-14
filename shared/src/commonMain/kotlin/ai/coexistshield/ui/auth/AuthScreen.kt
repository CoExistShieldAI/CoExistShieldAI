package ai.coexistshield.ui.auth

import ai.coexistshield.auth.AuthUser
import ai.coexistshield.auth.AuthService
import ai.coexistshield.auth.AuthValidator
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.launch

@Composable
fun AuthScreen(
    authService: AuthService,
    onAuthenticated: (AuthUser) -> Unit,
    onContinueAsGuest: () -> Unit
) {
    var isRegister by remember { mutableStateOf(false) }
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var confirm by remember { mutableStateOf("") }
    var submitted by remember { mutableStateOf(false) }
    var isLoading by remember { mutableStateOf(false) }
    var serverError by remember { mutableStateOf<String?>(null) }
    var infoMessage by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    val emailError = if (submitted) AuthValidator.validateEmail(email) else null
    val passwordError = if (submitted) AuthValidator.validatePassword(password) else null
    val confirmError = if (submitted && isRegister) AuthValidator.validateConfirm(password, confirm) else null

    fun resetPassword() {
        submitted = true
        serverError = null
        infoMessage = null
        if (AuthValidator.validateEmail(email) != null) return
        isLoading = true
        scope.launch {
            val result = authService.sendPasswordReset(email)
            isLoading = false
            result.fold(
                onSuccess = { infoMessage = "パスワード再設定用のメールを送信しました (登録済みの場合)" },
                onFailure = { serverError = it.message }
            )
        }
    }

    fun submit() {
        submitted = true
        serverError = null
        infoMessage = null
        if (AuthValidator.validateEmail(email) != null ||
            AuthValidator.validatePassword(password) != null ||
            (isRegister && AuthValidator.validateConfirm(password, confirm) != null)
        ) return
        isLoading = true
        scope.launch {
            val result = if (isRegister) authService.signUp(email, password) else authService.signIn(email, password)
            isLoading = false
            result.fold(
                onSuccess = onAuthenticated,
                onFailure = { serverError = it.message }
            )
        }
    }

    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Column(
            modifier = Modifier.widthIn(max = 420.dp).fillMaxWidth().padding(24.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text("🛡️ CoExist Shield AI", fontSize = 22.sp, fontWeight = FontWeight.Black)
            Text(
                if (isRegister) "Create New Account" else "Sign In",
                fontSize = 16.sp,
                fontWeight = FontWeight.Bold
            )

            OutlinedTextField(
                value = email,
                onValueChange = { email = it },
                label = { Text("Email Address") },
                singleLine = true,
                isError = emailError != null,
                supportingText = { emailError?.let { Text(it) } },
                modifier = Modifier.fillMaxWidth()
            )
            OutlinedTextField(
                value = password,
                onValueChange = { password = it },
                label = { Text("Password (8+ chars)") },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                isError = passwordError != null,
                supportingText = { passwordError?.let { Text(it) } },
                modifier = Modifier.fillMaxWidth()
            )
            if (isRegister) {
                OutlinedTextField(
                    value = confirm,
                    onValueChange = { confirm = it },
                    label = { Text("Confirm Password") },
                    singleLine = true,
                    visualTransformation = PasswordVisualTransformation(),
                    isError = confirmError != null,
                    supportingText = { confirmError?.let { Text(it) } },
                    modifier = Modifier.fillMaxWidth()
                )
            }

            serverError?.let {
                Text(it, color = MaterialTheme.colorScheme.error, fontSize = 13.sp)
            }

            infoMessage?.let {
                Text(it, color = MaterialTheme.colorScheme.primary, fontSize = 13.sp)
            }

            Button(
                onClick = { submit() },
                enabled = !isLoading,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(
                    when {
                        isLoading -> "Processing..."
                        isRegister -> "Register Account"
                        else -> "Sign In"
                    }
                )
            }
            if (!isRegister) {
                TextButton(onClick = { resetPassword() }, enabled = !isLoading) {
                    Text("パスワードを忘れた方")
                }
            }
            TextButton(
                onClick = {
                    isRegister = !isRegister
                    infoMessage = null
                    submitted = false
                    serverError = null
                },
                enabled = !isLoading
            ) {
                Text(if (isRegister) "Already have an account? Sign In" else "New to CoExist? Register")
            }
            TextButton(onClick = onContinueAsGuest, enabled = !isLoading) {
                Text("Continue as Guest (Demo Mode)")
            }
        }
    }
}
