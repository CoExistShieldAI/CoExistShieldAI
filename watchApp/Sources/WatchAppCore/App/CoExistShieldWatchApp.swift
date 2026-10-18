import SwiftUI

#if os(watchOS)
@main
public struct CoExistShieldWatchApp: App {
    @StateObject private var repository = WatchStateRepository.shared
    private let sessionManager = WatchSessionManager.shared

    public init() {
        sessionManager.startSession()
    }

    public var body: some Scene {
        WindowGroup {
            WatchHudView(repository: repository, sessionManager: sessionManager)
        }
    }
}
#endif
