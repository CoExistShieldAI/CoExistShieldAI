import Foundation
import Combine

/// Thread-safe singleton repository holding current HUD state for watchOS.
public final class WatchStateRepository: ObservableObject {
    public static let shared = WatchStateRepository()

    @Published public private(set) var hudState: WatchHudState

    private let lock = NSLock()

    public init(initialState: WatchHudState = WatchHudState()) {
        self.hudState = initialState
    }

    public func updateState(
        tier: AlertTier,
        distanceMeters: Double?,
        species: [WildlifeSpecies],
        isSirenPlaying: Bool
    ) {
        lock.lock()
        defer { lock.unlock() }

        let newState = WatchHudState(
            tier: tier,
            closestDistanceMeters: distanceMeters,
            targetSpecies: species,
            isSirenPlaying: isSirenPlaying,
            lastUpdatedMillis: Int64(Date().timeIntervalSince1970 * 1000)
        )

        DispatchQueue.main.async { [weak self] in
            self?.hudState = newState
        }
    }

    public func updateFromDictionary(_ dict: [String: Any]) {
        let newState = WatchHudState.from(dictionary: dict)
        lock.lock()
        defer { lock.unlock() }

        DispatchQueue.main.async { [weak self] in
            self?.hudState = newState
        }
    }

    public func silenceSiren() {
        lock.lock()
        defer { lock.unlock() }

        var current = hudState
        current.isSirenPlaying = false

        DispatchQueue.main.async { [weak self] in
            self?.hudState = current
        }
    }
}
