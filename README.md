# CoExist Shield AI

> **Autonomous AI-Driven Predictive Geofencing and Ecological Early-Warning Platform for Human-Wildlife Coexistence and Global One Health**

---

## 1. Executive Summary & Vision

**CoExist Shield AI** is an intelligent early-warning and autonomous incident response platform engineered to eliminate fatal human-wildlife encounters and safeguard biodiversity.

At the core of the platform is **Dynamic Predictive Geofencing**. Autonomous AI agents continuously ingest and correlate multi-source environmental and causal data streams to forecast the probability of dangerous wildlife appearances. Based on these predictive models, the system dynamically establishes geofences and delivers a **3-tier early warning notification system** via the **CoExist Shield Mobile Application**, ensuring citizens and field personnel receive actionable notice before entering hazardous zones.

In addition, the platform uniquely bridges **wildlife tracking** and **roadkill incident management**, transforming routine municipal carcass disposal into a vital global **"One Health"** safety barrier.

---

## 2. Core Philosophy & Strategic Value

### 2.1 Why Unify Wildlife Tracking and Roadkill Reporting?
Conventional approaches treat wildlife encounters and roadkill management as disconnected domains. CoExist Shield AI unifies them based on deep ecological and epidemiological connections:

1. **Ecological Corridors:** High-incidence roadkill zones pinpoint natural wildlife movement corridors and migration routes.
2. **Carcasses as Predatory Magnets (Leading Indicator of Attacks):** Decomposing carcasses left on roadways act as an open buffet, directly attracting apex predators (such as bears, tigers, and leopards) into proximity with human settlements. Roadkill data is a critical **leading indicator** of imminent human-wildlife conflict. Prompt carcass removal neutralizes the attractant before dangerous predators are lured in.
3. **Global One Health Barrier:**
   - **Japan:** Modernizes and streamlines carcass collection and road infrastructure management for municipal health and public works departments through AI-driven DX.
   - **APAC & Tropical Regions:** Serves as an autonomous surveillance net that rapidly identifies breeding grounds for infectious diseases and shuts down illegal, dangerous bushmeat supply chains, preventing the next zoonotic pandemic.

### 2.2 Addressing Root Causes: The Japan Bear Conflict Case Study
For decades, public consensus assumed that bears invade human settlements primarily because they starve in mountainous habitats. Modern empirical data—derived from the analysis of over 650 bears—disproves this assumption:

- **The Real Driver:** Bears enter towns predominantly drawn by **unharvested and abandoned persimmon (*kaki*) trees** and fruit groves in depopulated, aging rural communities. These neglected, nutrient-dense fruits act as an irresistible lure directly into human backyards.
- **Proactive AI Mitigation:** Rather than merely tracking bear coordinates after an encounter occurs, CoExist Shield AI cross-references regional vegetation, crop phenology, and municipal demographic data to evaluate the hazard score of abandoned fruit trees. The system charts optimized harvesting and clearing routes for local community action, extinguishing the root incentive for bears to encroach upon residential areas.

---

## 3. Target Scope & Fauna (Phase 1)

| Species | Primary Target Regions | Threat Profile & Interaction Dynamics |
| :--- | :--- | :--- |
| **Brown Bears & Asian Black Bears** (*Ursus arctos*, *Ursus thibetanus*) | Japan | Encroachment fueled by abandoned fruit trees and residential waste; pre-hibernation hyperphagia; roadkill attractants. |
| **Asian Elephants** (*Elephas maximus*) | Thailand, India, Sri Lanka | Crop raiding along agricultural borders; migration corridor fragmentation; railway/road collisions. |
| **Sumatran & Bengal Tigers** (*Panthera tigris sumatrae*, *P. t. tigris*) | Indonesia, India, Nepal | Habitat loss displacement (wildfires/deforestation); nocturnal village predation; attraction to livestock and roadkill. |

---

## 4. System Architecture & Core Capabilities

```
+-----------------------------------------------------------------------------------+
|                           Autonomous AI Agent Brain                               |
|                                                                                   |
|  +---------------------------+   +---------------------------------------------+  |
|  | Causal Web & API Streams  |   | Meteorological & Environmental Dynamics     |  |
|  | - Agriculture & Harvest   |   | - Precipitation & Snowpack Dynamics         |  |
|  | - Waste Management Cycles |   | - Temperature & Humidity (Decomposition)    |  |
|  | - FIRMS Thermal & Logging |   | - Moon Phase & Nocturnal Illumination       |  |
|  | - Social Media NLP (X/BS) |   | - Wind Direction & Scent Plumes             |  |
|  | - Dynamic Road Traffic    |   +---------------------------------------------+  |
|  +---------------------------+                          |                         |
|                 |                                       |                         |
|                 +-------------------+-------------------+                         |
|                                     |                                             |
|                                     v                                             |
|                 +---------------------------------------+                         |
|                 | Geospatial Probability Inference      |                         |
|                 | & Dynamic Geofence Generator Engine   |                         |
|                 +---------------------------------------+                         |
+-------------------------------------|---------------------------------------------+
                                      | Dynamic Hazard Polygons / Risk Scores
                                      v
+-----------------------------------------------------------------------------------+
|                        CoExist Shield Mobile Application                          |
|                                                                                   |
|  [ 3-Tier Geofence Alert HUD ]     [ Crowdsourced Ingestion ]   [ Roadkill Log ]  |
|  - Level 1: Notice (Early Prep)    - Direct Sightings & Attacks - Photo + GPS Log |
|  - Level 2: Warning (High Caution) - Traces (Scat, Prints, Claw)- Map Overlay     |
|  - Level 3: Danger (Evacuate Now)  - Offline Sync Capability    - Status Workflow |
+-----------------------------------------------------------------------------------+
```

### 4.1 CoExist Shield Mobile Application
1. **3-Tier Dynamic Geofence Warning HUD:**
   - **Level 1 (Notice / Green-Yellow):** Low-to-moderate probability zone. Informs users well in advance so they can stay alert, adjust travel, or prepare deterrents (e.g., bear bells, repellent).
   - **Level 2 (Warning / Amber):** High encounter likelihood. Advises detours and prompts heightened vigilance.
   - **Level 3 (Danger / Red):** Immediate threat or active apex predator presence. Sounds high-priority sirens/vibrations directing immediate evacuation.
2. **Crowdsourced Environmental & Encounter Logging:**
   - Real-time logging of visual sightings, close encounters, or physical damage with high-accuracy GPS coordinates and timestamps.
   - Forensic trace logging: Allows users and rangers to submit photos of tracks, scat, tree claw marks, and damaged crops.
3. **Roadkill Incident Logging & Live Map:**
   - Rapid camera capture with automated GPS tagging.
   - Interactive GIS map displaying confirmed roadkill locations, removal status, and correlated predator hazard zones.

### 4.2 Autonomous AI Agent: Multi-Source Causal Data Ingestion
The AI Agent continuously gathers and learns from five critical causal data streams available via public APIs and real-time web scraping:

1. **Agricultural Harvest & Crop Shipment Data:**
   - *Causality:* Fruit harvest seasons (persimmons, chestnuts, apples) in Japan and oil palm fruit harvesting peaks in Southeast Asia emit intense scents and yield agricultural waste, strongly attracting herbivores and opportunistic carnivores.
   - *Sources:* Government agricultural statistics APIs, municipal farming calendars, and wholesale market shipment feeds.
2. **Municipal Waste & Refuse Schedules:**
   - *Causality:* Periodic garbage collection days (notably organic waste) and unmanaged waste around tourist campgrounds. Highly intelligent animals (bears, monkeys) memorize weekly collection cycles (e.g., Tuesday/Friday routines).
   - *Sources:* Municipal open-data portals (JSON/CSV) and campsite reservation/utilization indices.
3. **Thermal Infrared Satellite & Deforestation Data:**
   - *Causality:* Slash-and-burn agriculture, wildfires, or extensive logging (e.g., solar farm clearing in mountain regions). Displaced apex predators are forced into adjacent human settlements within days to weeks.
   - *Sources:* NASA FIRMS (Fire Information for Resource Management System) API and NOAA satellite thermal anomalies.
4. **Social & Real-Time Crowdsourced NLP:**
   - *Causality:* Citizen reports on social media often precede official municipal warnings by hours or days ("Saw a bear near Route 4", "Dead deer blocking the lane").
   - *Sources:* X (formerly Twitter) API, Bluesky API, and local online bulletin boards. Large Language Models (LLMs) parse sentiment, extract wildlife entities, and automatically geocode incidents.
5. **Dynamic Traffic & Roadway Network Feeds:**
   - *Causality:* Spikes in regional road traffic during holiday seasons increase roadkill frequency and sever wildlife migratory corridors, triggering disorientation and panic-driven incursions into populated areas.
   - *Sources:* Google Maps Traffic API and national highway agency real-time congestion feeds.

### 4.3 Meteorological & Environmental Dynamics
The AI Agent models four crucial weather variables directly linked to animal behavior:

1. **Precipitation & Snowpack (Real-time & Forecast):**
   - *Causality:* Flooded mountain streams prevent crossing, forcing animals (elephants, bears) onto roads and cultivated plains. The first snowfall in Japan triggers intense pre-hibernation foraging (hyperphagia).
2. **Temperature & Humidity:**
   - *Causality:* Severe summer heat dries up mountain springs, driving wildlife toward irrigation canals and reservoirs. In tropical APAC, the onset of the rainy season drastically accelerates carcass decomposition, disease pathogen proliferation, and venomous snake activity.
3. **Moon Phase & Nocturnal Illumination:**
   - *Causality:* Nocturnal and crepuscular predators (tigers, bears) exhibit significantly higher village and road intrusion rates during **New Moon nights (total darkness)**.
4. **Wind Direction & Wind Speed:**
   - *Causality:* Wildlife rely heavily on olfaction. When prevailing winds blow from residential areas (dumpsters, unharvested orchards) toward forest ridges, predators follow scent plumes directly downhill.

---

## 5. Development & Project Roadmap

The engineering tasks and requirements are managed on GitHub Projects:
**[CoExistShieldAI Project Board #1](https://github.com/orgs/CoExistShieldAI/projects/1)**

All issues, pull requests, and git commit messages follow the **Conventional Commits** specification in English:
- `feat(...)`: New features and modules
- `fix(...)`: Bug fixes and behavioral patches
- `docs(...)`: Architectural, specification, and onboarding documentation
- `refactor(...)`: Code structuring without functional changes
- `test(...)`: Unit, integration, and simulation tests
