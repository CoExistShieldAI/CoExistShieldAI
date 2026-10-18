import SwiftUI

/// Heads-Up dynamic notification view for Apple Watch hazard alerts.
public struct WatchNotificationView: View {
    public let tier: AlertTier
    public let distanceMeters: Double?
    public let message: String

    public init(
        tier: AlertTier = .tier2Warning,
        distanceMeters: Double? = nil,
        message: String = "Wildlife incursion detected"
    ) {
        self.tier = tier
        self.distanceMeters = distanceMeters
        self.message = message
    }

    public var body: some View {
        VStack(spacing: 8) {
            HStack(spacing: 6) {
                Text(tier.icon)
                    .font(.title3)
                Text(tier.displayName)
                    .font(.system(size: 11, weight: .black, design: .monospaced))
                    .foregroundColor(tier.tacticalColor)
            }

            Text(message)
                .font(.system(size: 12, weight: .medium))
                .foregroundColor(.white)
                .multilineTextAlignment(.center)

            if let distance = distanceMeters {
                Text("\(Int(distance))m AWAY")
                    .font(.system(size: 14, weight: .black, design: .rounded))
                    .foregroundColor(Color(red: 0xF1 / 255.0, green: 0xF5 / 255.0, blue: 0xF9 / 255.0))
            }
        }
        .padding(8)
    }
}
