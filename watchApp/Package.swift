// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "WatchApp",
    platforms: [
        .watchOS(.v9),
        .iOS(.v16),
        .macOS(.v13)
    ],
    products: [
        .library(
            name: "WatchAppCore",
            targets: ["WatchAppCore"]
        )
    ],
    targets: [
        .target(
            name: "WatchAppCore",
            path: "Sources/WatchAppCore"
        ),
        .testTarget(
            name: "WatchAppCoreTests",
            dependencies: ["WatchAppCore"],
            path: "Tests/WatchAppCoreTests"
        )
    ]
)
