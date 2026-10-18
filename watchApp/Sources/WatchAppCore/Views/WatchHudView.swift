import SwiftUI

/// Main wrist-worn HUD screen for Apple Watch.
/// Displays high-contrast tactical telemetry and quick response actions.
public struct WatchHudView: View {
    @ObservedObject public var repository: WatchStateRepository
    public let sessionManager: WatchSessionManager

    @State private var showActionOverlay = false

    public init(
        repository: WatchStateRepository = .shared,
        sessionManager: WatchSessionManager = .shared
    ) {
        self.repository = repository
        self.sessionManager = sessionManager
    }

    public var body: some View {
        let state = repository.hudState

        GeometryReader { geometry in
            let availableSize = min(geometry.size.width, geometry.size.height)
            let meterSize = max(150.0, availableSize * 0.96)

            ZStack {
                Color.black.ignoresSafeArea()

                // Main Modern Tactical Meter
                ModernTacticalMeterView(
                    tier: state.tier,
                    distanceMeters: state.closestDistanceMeters,
                    species: state.targetSpecies,
                    meterSize: meterSize
                )
                .contentShape(Rectangle())
                .onTapGesture {
                    // Tap to toggle quick tactical actions during threats
                    if state.tier != .safe || state.isSirenPlaying {
                        withAnimation(.spring()) {
                            showActionOverlay.toggle()
                        }
                    }
                }

                // Emergency Action Overlay (Active when tapped or during Siren)
                if showActionOverlay || state.isSirenPlaying {
                    Color.black.opacity(0.85)
                        .ignoresSafeArea()
                        .transition(.opacity)

                    VStack(spacing: 8) {
                        Text(state.isSirenPlaying ? "🚨 SIREN ACTIVE" : "TACTICAL ACTIONS")
                            .font(.system(size: 11, weight: .black, design: .monospaced))
                            .foregroundColor(Color(red: 0xEF / 255.0, green: 0x44 / 255.0, blue: 0x44 / 255.0))
                            .multilineTextAlignment(.center)

                        // Silence / Acknowledge Button
                        Button(action: {
                            sessionManager.sendDismissAction()
                            withAnimation {
                                showActionOverlay = false
                            }
                        }) {
                            Text("SILENCE / ACK")
                                .font(.system(size: 10, weight: .bold, design: .rounded))
                                .foregroundColor(.white)
                                .frame(maxWidth: .infinity)
                                .padding(.vertical, 6)
                                .background(Color(red: 0x1E / 255.0, green: 0x29 / 255.0, blue: 0x3B / 255.0))
                                .cornerRadius(8)
                        }
                        .buttonStyle(PlainButtonStyle())
                        .frame(maxWidth: availableSize * 0.8)

                        // SOS Evacuation Dispatch Button (Tier 3 Danger)
                        if state.tier == .tier3Danger {
                            Button(action: {
                                sessionManager.sendSosAction()
                                withAnimation {
                                    showActionOverlay = false
                                }
                            }) {
                                Text("EVACUATE SOS")
                                    .font(.system(size: 10, weight: .black, design: .rounded))
                                    .foregroundColor(.white)
                                    .frame(maxWidth: .infinity)
                                    .padding(.vertical, 6)
                                    .background(Color(red: 0xDC / 255.0, green: 0x26 / 255.0, blue: 0x26 / 255.0))
                                    .cornerRadius(8)
                            }
                            .buttonStyle(PlainButtonStyle())
                            .frame(maxWidth: availableSize * 0.8)
                        }

                        // Dismiss overlay hint
                        Text("Tap to close")
                            .font(.system(size: 8.5, weight: .medium))
                            .foregroundColor(.gray)
                            .padding(.top, 2)
                    }
                    .padding(12)
                    .contentShape(Rectangle())
                    .onTapGesture {
                        withAnimation {
                            showActionOverlay = false
                        }
                    }
                }
            }
        }
    }
}
