import SwiftUI

/// High-precision, modern tactical HUD gauge matching the cross-platform ModernTacticalMeter presentation:
/// - Ambient radial glow with threat heartbeat pulse
/// - Outer circular track with 18% opacity
/// - Dynamic proximity arc with round cap (0° - 360°)
/// - Inner dashed tactical radar ring (dash 3, 6)
/// - Focused, high-contrast central typography for round & rectangular watch bezels
public struct ModernTacticalMeterView: View {
    public let tier: AlertTier
    public let distanceMeters: Double?
    public let species: [WildlifeSpecies]
    public let meterSize: CGFloat

    @State private var pulseScale: CGFloat = 1.0
    @State private var pulseAlpha: Double = 0.14
    @State private var animatedFillRatio: Double = 1.0

    public init(
        tier: AlertTier,
        distanceMeters: Double?,
        species: [WildlifeSpecies] = [],
        meterSize: CGFloat = 176.0
    ) {
        self.tier = tier
        self.distanceMeters = distanceMeters
        self.species = species
        self.meterSize = meterSize
    }

    public var body: some View {
        let tacticalColor = tier.tacticalColor
        let strokeWidth = max(7.0, meterSize * 0.048)
        let outerDiameter = meterSize - strokeWidth
        let innerDiameter = outerDiameter * 0.83 // Aligned with Compose r=78 / r=94 ratio

        ZStack {
            // 1. Ambient Radial Threat Glow
            Circle()
                .fill(
                    RadialGradient(
                        gradient: Gradient(colors: [
                            tacticalColor.opacity(pulseAlpha),
                            tacticalColor.opacity(0.04),
                            Color.clear
                        ]),
                        center: .center,
                        startRadius: 0,
                        endRadius: meterSize * 0.55
                    )
                )
                .frame(width: meterSize, height: meterSize)

            // 2. Outer Track Ring
            Circle()
                .stroke(tacticalColor.opacity(0.18), lineWidth: strokeWidth)
                .frame(width: outerDiameter, height: outerDiameter)

            // 3. Dynamic Proximity Arc (Clockwise from 12 o'clock / -90°)
            Circle()
                .trim(from: 0.0, to: CGFloat(animatedFillRatio))
                .stroke(
                    tacticalColor,
                    style: StrokeStyle(
                        lineWidth: strokeWidth,
                        lineCap: tier == .safe ? .square : .round
                    )
                )
                .rotationEffect(.degrees(-90))
                .frame(width: outerDiameter, height: outerDiameter)

            // 4. Inner Dashed Tactical Radar Ring
            Circle()
                .stroke(
                    tacticalColor.opacity(0.35),
                    style: StrokeStyle(lineWidth: 1.8, dash: [3, 6])
                )
                .frame(width: innerDiameter, height: innerDiameter)

            // 5. Central Tactical Information Orb
            VStack(spacing: 1) {
                // Tier Icon
                Text(tier.icon)
                    .font(.system(size: meterSize < 170 ? 22 : 28))

                // Tier Name
                Text(tier.displayName)
                    .font(.system(size: meterSize < 170 ? 10 : 11, weight: .black, design: .monospaced))
                    .foregroundColor(tacticalColor)
                    .tracking(0.8)
                    .multilineTextAlignment(.center)

                // Distance or Status
                if let distance = distanceMeters {
                    Text("\(Int(distance))m")
                        .font(.system(size: meterSize < 170 ? 18 : 22, weight: .black, design: .rounded))
                        .foregroundColor(Color(red: 0xF1 / 255.0, green: 0xF5 / 255.0, blue: 0xF9 / 255.0))

                    Text("ESTIMATED PROXIMITY")
                        .font(.system(size: meterSize < 170 ? 6.5 : 7.5, weight: .bold, design: .monospaced))
                        .foregroundColor(Color(red: 0x94 / 255.0, green: 0xA3 / 255.0, blue: 0xB8 / 255.0))
                        .tracking(0.6)
                } else {
                    Text("ALL SECURE")
                        .font(.system(size: meterSize < 170 ? 13 : 15, weight: .heavy, design: .rounded))
                        .foregroundColor(Color(red: 0xF1 / 255.0, green: 0xF5 / 255.0, blue: 0xF9 / 255.0))

                    Text("CONTINUOUS SENSING")
                        .font(.system(size: meterSize < 170 ? 6.5 : 7.5, weight: .bold, design: .monospaced))
                        .foregroundColor(Color(red: 0x94 / 255.0, green: 0xA3 / 255.0, blue: 0xB8 / 255.0))
                        .tracking(0.6)
                }

                // Target Species Badges (if any)
                if !species.isEmpty {
                    Text(species.map { $0.displayName.uppercased() }.joined(separator: " • "))
                        .font(.system(size: 8.5, weight: .bold, design: .monospaced))
                        .foregroundColor(tacticalColor)
                        .lineLimit(1)
                        .padding(.top, 1)
                }
            }
            .padding(.horizontal, 12)
        }
        .frame(width: meterSize, height: meterSize)
        .onAppear {
            updateFillRatio()
            startPulseAnimation()
        }
        .onChange(of: tier) { _ in
            updateFillRatio()
            startPulseAnimation()
        }
        .onChange(of: distanceMeters) { _ in
            updateFillRatio()
        }
    }

    private func updateFillRatio() {
        let state = WatchHudState(tier: tier, closestDistanceMeters: distanceMeters)
        withAnimation(.easeInOut(duration: 0.6)) {
            animatedFillRatio = state.proximityFillRatio
        }
    }

    private func startPulseAnimation() {
        let isDanger = (tier == .tier3Danger)
        let duration = isDanger ? 0.45 : 1.2
        let targetAlpha = isDanger ? 0.40 : 0.22

        withAnimation(
            Animation
                .easeInOut(duration: duration)
                .repeatForever(autoreverses: true)
        ) {
            pulseAlpha = targetAlpha
        }
    }
}
