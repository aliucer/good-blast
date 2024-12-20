# Good Blast API

This repository provides the backend for the **Good Blast** casual Match 3 game, focusing on user progress, daily tournaments, leaderboards, and tournament rewards. It is built with **Go**, uses **DynamoDB**  for persistent storage, **Redis** for caching leaderboard data, and is deployed on **Fly.io**.

## API Design

The API is structured around **RESTful endpoints**, using a layered architecture:

- **Handlers (Presentation Layer):**  
  [Gin](https://github.com/gin-gonic/gin)-based HTTP handlers under `api/handlers/` process incoming requests, validate input, and return JSON responses.
  
- **Services (Business Logic Layer):**  
  The `services/` directory defines core business logic. Key services include:  
  - **UserService:** Manages user creation, retrieval, and progress updates.  
  - **TournamentService:** Handles daily tournament lifecycle (start, end, user entries, scoring, and rewards).  
  - **LeaderboardService:** Retrieves global, country-level, and tournament-specific leaderboards.
  
- **Database Layer:**  
  Implemented in `database/` using a `DatabaseInterface`. **Amazon DynamoDB** serves as the persistent storage.  
  DynamoDB tables:
  - **Users Table:** User records keyed by `userId`. GSIs enable global and country-based level sorting.  
    - **GlobalLevelIndex:** (globalPK, level) for global leaderboard.  
    - **CountryLevelIndex:** (country, level) for country-specific leaderboard.
  - **Tournaments Table:** One record per daily tournament keyed by `tournamentId` (formatted date).
  - **TournamentEntries Table:** Entries keyed by (tournamentId, userId) with a `GroupScoreIndex` for leaderboards within groups.

- **Caching (Redis):**  
  Leaderboard queries are cached in Redis for short periods (e.g., 60 seconds) to reduce DynamoDB load and improve response times under heavy read conditions.

## Key Features

### User Management
- **Create Users:** Each new user starts at level 1 with 1000 coins.  
- **Update Progress:** Users gain 100 coins per level advancement.

### Tournament Operations
- **Automatic Daily Tournaments:**  
  A new tournament starts at **00:00 UTC** daily. The previous day’s tournament ends at **23:59 UTC**.
- **Entry Requirements:**  
  Users must be level ≥10 and pay 500 coins to enter. They join 35-person groups.
- **Scoring & Rewards:**  
  Scores increment as users progress. When a tournament ends, rewards are distributed based on rank within the user’s group:
  - 1st place: 5000 coins
  - 2nd place: 3000 coins
  - 3rd place: 2000 coins
  - 4th–10th places: 1000 coins
- **Daily Rotation:**  
  Automatically end yesterday’s tournament and start a new one at midnight.

### Leaderboards
- **Global Leaderboard:** Top 1000 users by level.  
- **Country Leaderboard:** Top 1000 users by level within a specific country.  
- **Tournament Leaderboard:** Rankings and scores within a tournament group.  
- **Caching:** Redis reduces response latency and DynamoDB reads.

### Cron Integration (Automated Management)
You can set up a cron job (or a scheduled task) to:
- **End Yesterday’s Tournament:** `PUT /tournaments/end/{yesterdaysDate}`
- **Start Today’s Tournament:** `POST /tournaments/start`
at midnight UTC daily. This ensures a seamless daily tournament cycle.

## Used Technologies
- **Language:** Go  
- **HTTP Framework:** Gin  
- **Database:** Amazon DynamoDB (configured in `eu-north-1`)  
- **Caching:** Redis  
- **Cloud Hosting:** Fly.io for deployment

## Deployment and Running

### DynamoDB Setup

### Redis
Redis runs inside the same container, as specified by the Dockerfile and `start.sh` script.

### Building and Deploying on Fly.io

#### Prerequisites:
- **Fly CLI:** Installed and configured.
- **AWS Credentials:** If running locally or specifying endpoints.
- **Environment Variables:**
  - `DYNAMODB_REGION=eu-north-1`
  - `USERS_TABLE=Users`
  - `TOURNAMENTS_TABLE=Tournaments`
  - `TOURNAMENT_ENTRIES_TABLE=TournamentEntries`
  
#### Build and Run Locally:
```bash
docker build -t good-blast-real .
docker run -p 8080:8080 \
  -e DYNAMODB_REGION=eu-north-1 \
  -e USERS_TABLE=Users \
  -e TOURNAMENTS_TABLE=Tournaments \
  -e TOURNAMENT_ENTRIES_TABLE=TournamentEntries \
  good-blast-real ```

### Access the API at: [http://localhost:8080](http://localhost:8080) 
### Deploy to Fly.io: 
```bash 
fly deploy ```

Once deployed, the application is accessible at: [https://good-blast-real.fly.dev/](https://good-blast-real.fly.dev/) 
### Automated Daily Tournaments with Cron Integrate a cron-like mechanism on Fly.io or use external services. 

At midnight UTC: - `PUT /tournaments/end/{yesterdaysDate}` - `POST /tournaments/start` 
### Testing 
- **Unit Tests:** In `services/` for `UserService`, `TournamentService`, and `LeaderboardService. 
- **Mocked Database:** Tests run with a mock `DatabaseInterface`, no real DynamoDB calls needed. Run Tests: ```bash go test ./... ``` 
This README describes the **backend architecture**, **database structure**, **caching**, and **deployment** process. `


