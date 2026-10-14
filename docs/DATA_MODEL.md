# Data Model — Canonical Enums

Single source of truth: the Kotlin models in `shared/src/commonMain/kotlin/ai/coexistshield/model/`.
Issue text, Gemini JSON schemas, and `firestore.rules` must use these values exactly.

| Enum | Values |
| :--- | :--- |
| `UserRole` | `CITIZEN`, `OFFICIAL`, `ADMIN` |
| `AlertTier` | `SAFE`, `TIER_1_NOTICE`, `TIER_2_WARNING`, `TIER_3_DANGER` |
| `WildlifeSpecies` | `TIGER`, `ELEPHANT`, `BEAR`, `WILD_BOAR`, `DEER`, `MONKEY`, `UNKNOWN` |
| `ReportCategory` | `WILDLIFE_SIGHTING`, `PHYSICAL_TRACE`, `ROADKILL`, `LIVESTOCK_DEPREDATION`, `ATTRACTANT_FRUIT_TREE` |
| `VerificationStatus` | `PENDING`, `VERIFIED`, `REJECTED` |
| `DecayState` | `FRESH`, `BLOATED`, `DECOMPOSING`, `SKELETAL` |
| `ClearanceStatus` | `REPORTED`, `DISPATCHED`, `CLEARED` |
| `FruitLoadLevel` | `SPARSE`, `MODERATE`, `ABUNDANT` |
| `RipenessStage` | `UNRIPE_GREEN`, `RIPE`, `OVERRIPE_FALLING` |
| `TreeManagementStatus` | `UNMANAGED`, `SCHEDULED_REMOVAL`, `PRUNED`, `FELLED` |
| `CropGrowthStage` | `VEGETATIVE`, `RIPENING`, `READY_FOR_HARVEST`, `HARVESTED_CLEARED` |
| `Region` (rules/claims) | `JAPAN`, `APAC` |

## Conventions
- Firestore `location` is a `GeoPoint` (`{_lat,_long}`); the Kotlin `GeoLocation(latitude, longitude, altitude?, geohash?)` maps to it.
- Geohash precision: ≤5 characters for anything public (`kaki_trees_public`).
- `geofences.polygonGeoJson` accepts `Polygon | MultiPolygon` (RFC 7946).
- Sample data: Morioka City is `municipalityCode=03201` (~39.70N, 141.15E).
- **Field Media Attachments (Google Cloud Storage):**
  - Storage Bucket: `corald-coexistshield-media` (or `GCS_MEDIA_BUCKET` env).
  - Object Key Pattern: `uploads/YYYYMMDD_HHMMSS_<rand4>_<sanitized_name>.<ext>`.
  - Max limits: Up to 5 images per report, maximum 5 MB per file.
  - Multi-photo field `photo_urls: string[]` is canonical across Wildlife, Roadkill, and Environmental records.
  - For backward compatibility with legacy single-photo consumers, Roadkill `photo_url` mirrors `photo_urls[0]`.

## Known divergences still to resolve
- `hazardTier` (`LOW|MEDIUM|HIGH`, used by `kaki_trees_public`) has no Kotlin enum yet.
- `Region` has no Kotlin enum yet; `OFFICIAL_JAPAN` / `OFFICIAL_APAC` in #28 are expressed as `role=OFFICIAL` + `region` claim.
- `CropType` and `FruitType` differ (crops vs. fruit trees); registration routing between `kaki_trees` and `crop_attractants` is undecided.
