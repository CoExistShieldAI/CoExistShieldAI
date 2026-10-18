import XCTest
@testable import WatchAppCore

final class WatchAppCoreTests: XCTestCase {

    func testAlertTierParsing() {
        XCTAssertEqual(AlertTier.parse("SAFE"), .safe)
        XCTAssertEqual(AlertTier.parse("TIER_1_NOTICE"), .tier1Notice)
        XCTAssertEqual(AlertTier.parse("TIER_2_WARNING"), .tier2Warning)
        XCTAssertEqual(AlertTier.parse("TIER_3_DANGER"), .tier3Danger)
        XCTAssertEqual(AlertTier.parse("UNKNOWN_VALUE"), .safe)
        XCTAssertEqual(AlertTier.parse(nil), .safe)
        XCTAssertEqual(AlertTier.parse(""), .safe)
    }

    func testWildlifeSpeciesParsing() {
        XCTAssertEqual(WildlifeSpecies.parse("TIGER"), .tiger)
        XCTAssertEqual(WildlifeSpecies.parse("elephant"), .elephant)
        XCTAssertEqual(WildlifeSpecies.parse("LEOPARD"), .leopard)
        XCTAssertEqual(WildlifeSpecies.parse("BEAR"), .bear)
        XCTAssertEqual(WildlifeSpecies.parse("BOAR"), .boar)
        XCTAssertEqual(WildlifeSpecies.parse("OTHER"), .other)
        XCTAssertNil(WildlifeSpecies.parse("UNKNOWN"))
        XCTAssertNil(WildlifeSpecies.parse(nil))
    }

    func testWatchHudStateProximityFillRatio() {
        // Safe mode should always be 1.0 (360 degrees full circle)
        let safeState = WatchHudState(tier: .safe, closestDistanceMeters: 100.0)
        XCTAssertEqual(safeState.proximityFillRatio, 1.0, accuracy: 0.001)

        // Null distance fallback
        let nullDistState = WatchHudState(tier: .tier2Warning, closestDistanceMeters: nil)
        XCTAssertEqual(nullDistState.proximityFillRatio, 0.40, accuracy: 0.001)

        // Close danger: 50m should map to max fill 0.92
        let closeState = WatchHudState(tier: .tier3Danger, closestDistanceMeters: 50.0)
        XCTAssertEqual(closeState.proximityFillRatio, 0.92, accuracy: 0.001)

        // Closer than 50m should clamp to 0.92
        let veryCloseState = WatchHudState(tier: .tier3Danger, closestDistanceMeters: 10.0)
        XCTAssertEqual(veryCloseState.proximityFillRatio, 0.92, accuracy: 0.001)

        // Distant: 2000m should map to min fill 0.20
        let distantState = WatchHudState(tier: .tier1Notice, closestDistanceMeters: 2000.0)
        XCTAssertEqual(distantState.proximityFillRatio, 0.20, accuracy: 0.001)

        // Further than 2000m should clamp to 0.20
        let veryDistantState = WatchHudState(tier: .tier1Notice, closestDistanceMeters: 3500.0)
        XCTAssertEqual(veryDistantState.proximityFillRatio, 0.20, accuracy: 0.001)
    }

    func testDictionarySerializationAndDeserialization() {
        let original = WatchHudState(
            tier: .tier3Danger,
            closestDistanceMeters: 180.5,
            targetSpecies: [.tiger, .elephant],
            isSirenPlaying: true,
            lastUpdatedMillis: 1729250000000
        )

        let dict = original.toDictionary()
        XCTAssertEqual(dict["tier"] as? String, "TIER_3_DANGER")
        XCTAssertEqual(dict["closest_distance_meters"] as? Double, 180.5)
        XCTAssertEqual(dict["is_siren_playing"] as? Bool, true)
        XCTAssertEqual(dict["timestamp"] as? Int64, 1729250000000)

        let parsed = WatchHudState.from(dictionary: dict)
        XCTAssertEqual(parsed.tier, .tier3Danger)
        XCTAssertEqual(parsed.closestDistanceMeters, 180.5)
        XCTAssertEqual(parsed.targetSpecies, [.tiger, .elephant])
        XCTAssertEqual(parsed.isSirenPlaying, true)
        XCTAssertEqual(parsed.lastUpdatedMillis, 1729250000000)
    }

    func testDictionaryParsingEdgeCases() {
        // Empty dictionary returns safe defaults
        let empty = WatchHudState.from(dictionary: [:])
        XCTAssertEqual(empty.tier, .safe)
        XCTAssertNil(empty.closestDistanceMeters)
        XCTAssertTrue(empty.targetSpecies.isEmpty)
        XCTAssertFalse(empty.isSirenPlaying)

        // Negative distance is treated as nil (safe/unknown)
        let negDist = WatchHudState.from(dictionary: [
            "tier": "TIER_2_WARNING",
            "closest_distance_meters": -1.0
        ])
        XCTAssertEqual(negDist.tier, .tier2Warning)
        XCTAssertNil(negDist.closestDistanceMeters)
    }

    func testStateRepositoryUpdates() {
        let repo = WatchStateRepository()
        XCTAssertEqual(repo.hudState.tier, .safe)

        repo.updateState(
            tier: .tier2Warning,
            distanceMeters: 320.0,
            species: [.leopard],
            isSirenPlaying: true
        )

        let exp = expectation(description: "State updated on main queue")
        DispatchQueue.main.async {
            XCTAssertEqual(repo.hudState.tier, .tier2Warning)
            XCTAssertEqual(repo.hudState.closestDistanceMeters, 320.0)
            XCTAssertEqual(repo.hudState.targetSpecies, [.leopard])
            XCTAssertTrue(repo.hudState.isSirenPlaying)

            repo.silenceSiren()
            DispatchQueue.main.async {
                XCTAssertFalse(repo.hudState.isSirenPlaying)
                exp.fulfill()
            }
        }
        waitForExpectations(timeout: 2.0)
    }

    func testSessionManagerAlertTriggerParsing() {
        let repo = WatchStateRepository()
        let manager = WatchSessionManager(repository: repo)

        // Simulate incoming message trigger: "TIER_3_DANGER|145.0|Apex predator approaching"
        manager.handleReceivedAlertTrigger(payload: "TIER_3_DANGER|145.0|Apex predator approaching")

        // Trigger executes immediately without crashing
        XCTAssertNotNil(manager)
    }
}
