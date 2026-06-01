# NOANT Enterprise — v2.0 (OmAgent Custom Edition)
### Autonomous & Human-in-the-Loop AI Customer Support Platform
*Built in Nigeria. For the World.*

---

## 📖 Table of Contents
1. [Overview & Visual Identity](#-overview--visual-identity)
2. [Architecture & System Topology](#-architecture--system-topology)
3. [AI Orchestration & "In-the-Round" Communication Flow](#-ai-orchestration--in-the-round-communication-flow)
4. [Core Subsystems & Features Deep Dive](#-core-subsystems--features-deep-dive)
5. [Database Schema & TiDB Storage Topology](#-database-schema--tidb-storage-topology)
6. [Premium Accomplishments & Enhancements](#-premium-accomplishments--enhancements)
7. [Installation & Local Deployment Guide](#-installation--local-deployment-guide)
8. [Comprehensive API Directory](#-comprehensive-api-directory)
9. [License & Development Credits](#-license--development-credits)

---

## 🎨 Overview & Visual Identity

**NOANT Enterprise** is a state-of-the-art, high-throughput, multi-tenant customer relationship and messaging platform designed to support high-performance businesses across Nigeria, Africa, and globally.

The system features a **hybrid automation model**—balancing high-confidence AI answering capabilities with seamless, real-time human escalation. 

### 👁️ The Visual Branding (Logo)
The branding identity of the system is modeled after an advanced conversation-loop icon:
* **The Outer Dashed Ring**: Symbolizes the multi-channel messaging endpoints (WhatsApp, Telegram, Facebook, Web Widget) spinning in perpetual synchrony.
* **The Solid Black Core**: Represents the central AI Brain container.
* **The Three White Dots**: Symbolize message propagation, growing in scale as intelligence accumulates.

> [!NOTE]
> The browser favicon has been fully engineered from this high-resolution logo, package-generating standard transparent `.ico`, `.png` assets (including modern responsive vector `favicon.svg` and `apple-touch-icon.png` for mobile devices) for 100% device compatibility.

---

## 🏗️ Architecture & System Topology

The platform is designed around a decoupled microservice-compatible structure, enforcing a strict separation of concerns, multi-tenant workspace isolation, and real-time state synchronization.

```mermaid
graph TD
    %% Clients
    subgraph Clients ["Client Gateways"]
        WA[WhatsApp App / Twilio]
        TG[Telegram Bot API]
        FB[Facebook Messenger API]
        WB[Web Widget Custom JS]
    end

    %% Routing Layer
    subgraph Edge ["Network & Security Layer"]
        R_PROXY[Reverse Proxy / CORS Guard]
        JWT_AUTH[JWT Auth Validator]
        RATE_LIM[Upstash Redis Rate Limiter]
    end

    %% Application Backend
    subgraph Backend ["Go 1.22 App Service (Gin)"]
        H_CHATS[ChatHandler]
        H_INTEG[IntegrationHandler]
        H_TRAIN[TrainingHandler]
        H_ANALY[AnalyticsHandler]
        
        S_INTEG[IntegrationService]
        S_WIDGET[WidgetService]
        S_CHAT[ChatService]
        
        AI_BRAIN[Groq Llama 3.3 AI Brain]
        WS_HUB[WebSocket Hub / Broadcast Server]
    end

    %% Persistence Layer
    subgraph Storage ["Distributed Persistence"]
        DB_TIDB[TiDB Cloud SQL Distributed Cluster]
        CACHE_RED[Upstash Cache & Memory Store]
    end

    %% Frontend App
    subgraph UI ["React SPA Frontend (Vite)"]
        RT_CONN[Network Context & Banners]
        WC_CTX[Widget Config Context]
        RT_CHATS[Chats Inbox View]
        ANALY_DASH[Overview Analytics]
    end

    %% Connections
    WA --> R_PROXY
    TG --> R_PROXY
    FB --> R_PROXY
    WB --> R_PROXY

    R_PROXY --> JWT_AUTH
    JWT_AUTH --> RATE_LIM
    RATE_LIM --> H_CHATS
    RATE_LIM --> H_INTEG
    RATE_LIM --> H_TRAIN
    RATE_LIM --> H_ANALY

    H_CHATS --> S_CHAT
    H_INTEG --> S_INTEG
    H_TRAIN --> S_CHAT
    H_ANALY --> S_CHAT

    S_CHAT --> AI_BRAIN
    S_CHAT --> WS_HUB
    S_WIDGET --> DB_TIDB

    S_CHAT --> DB_TIDB
    S_CHAT --> CACHE_RED

    WS_HUB -.->|WebSocket Real-time Feed| UI
    UI -->|API Requests| R_PROXY
```

---

## 🔄 AI Orchestration & "In-the-Round" Communication Flow

The "In-the-Round" orchestration describes how a message from an external user is safely consumed by the AI, verified against security and anti-hallucination thresholds, and either auto-replied or escalated to a live agent, closing the feedback loop through training.

```mermaid
sequenceDiagram
    autonumber
    actor Customer as WhatsApp / Widget User
    participant WH as Webhooks / Backend Handler
    participant AI as AI Brain & Llama 3.3
    participant DB as TiDB Database
    participant WS as WebSocket Broadcast Hub
    participant Client as React Dashboard (Agent View)
    actor Agent as Support Agent

    Customer->>WH: Sends message (e.g. "Do you ship to Lagos?")
    WH->>DB: Logs incoming message as unread
    WH->>WS: Broadcasts `new_message` event
    WS->>Client: Triggers real-time conversation list refresh (Last Message / Unread count update)
    
    WH->>AI: Prompts AI Brain with Query & Context
    activate AI
    AI->>DB: Semantic database search matching QA pairs
    DB-->>AI: Returns closest matching candidate (Confidence score)
    
    alt Confidence Score >= 0.70 (High Confidence Answering)
        AI-->>WH: Generates high-confidence response
        WH->>DB: Inserts AI message into `messages`
        WH->>WS: Broadcasts `new_message` to websocket
        WS->>Client: Automatically displays AI's answer in inbox
        WH->>Customer: Delivers reply back to customer chat
    else Confidence Score < 0.70 (Anti-Hallucination Guardrail & Escalation)
        AI-->>WH: Flags Query as Ambiguous / Unknown
        deactivate AI
        WH->>DB: Updates Conversation Status to 'escalated'
        WH->>DB: Inserts task in `unknown_questions` queue
        WH->>DB: Inserts preconfigured fallback warning text in `messages`
        WH->>WS: Broadcasts `unknown_question` & `integration_update`
        WS->>Client: Triggers visual pulse / alert, shows unread escalated badge
        WH->>Customer: Replies with: "I'm passing you to a human agent..."
        
        Agent->>Client: Selects conversation thread
        Client->>WH: GET /chats/conversations/:id
        WH->>DB: Calls `MarkRead` SQL, updates unread to 0
        DB-->>Client: Returns conversation thread & messages
        Agent->>WH: Clicks 'Takeover' conversation
        WH->>DB: Sets `taken_over_by` as Agent ID
        
        Agent->>Client: Resolves ticket & submits corrected answer in training form
        Client->>WH: POST /training/unknown-questions/:id/train
        WH->>DB: Inserts new QA pair, updates unknown_question status to 'trained'
        WH-->>Agent: Training complete! Knowledge Base instantly updated.
    end
```

---

## ⚙️ Core Subsystems & Features Deep Dive

### 1. The Anti-Hallucination Guardrail & Semantic Matcher
The Go backend implements a LangChain-style orchestration layer. When a user queries the bot:
* It searches the `qa_pairs` table using semantic full-text search.
* A strict confidence boundary of **`0.70`** is enforced.
* Under `0.70`, the system rejects the AI generation, logs the gap, triggers human escalation via real-time WebSocket signals, and saves the question in the dashboard's "Knowledge Gap" training view.

### 2. Live Platform Health Monitoring
To ensure 100% uptime for critical integrations:
* A background ticker routine in `main.go` performs diagnostic checks of Facebook, Instagram, WhatsApp, and Telegram integration channels every 5 minutes.
* Status flags are written back to the database, ensuring agents receive instant visual reports when channel configurations become invalid.

### 3. Role-Based Tenancy & Data Isolation
Security is structured for multi-agent support:
* **Tenant Isolation**: Every chat message, analytics query, and Q&A pair is bound directly to a filtered `user_id` context.
* **Role Heirarchy**: Owners and Admins manage integrations, configurations, and team invites, while Agents focus entirely on reading, resolving, and training conversations.

### 4. Direct Naira Payment Integrations
A built-in checkout gateway powered by **Polar** allows local Nigerian and international businesses to manage their subscriptions in Naira (NGN):
* Webhook endpoints process subscription upgrades, cancellations, and renewals.
* Features like max active team members, response limits, and connected channel metrics scale dynamically depending on the tier.

---

## 🗄️ Database Schema & TiDB Storage Topology

Our migrations are built for **TiDB Cloud** (distributed SQL clustering), optimizing read performance for real-time customer messaging.

### Database ER Model Summary
```
┌─────────────────────────────────┐       ┌─────────────────────────────────┐
│          users                  │       │          conversations          │
├─────────────────────────────────┤       ├─────────────────────────────────┤
│ id (PK)         VARCHAR(36)     │◄──┐   │ id (PK)         VARCHAR(36)     │◄──┐
│ email           VARCHAR(255)    │   │   │ user_id (FK)    VARCHAR(36)     │   │
│ password_hash   VARCHAR(255)    │   │   │ customer_name   VARCHAR(100)    │   │
│ first_name      VARCHAR(100)    │   │   │ channel         ENUM            │   │
│ last_name       VARCHAR(100)    │   │   │ status          ENUM            │   │
│ role            ENUM            │   │   │ intent          ENUM            │   │
│ plan_id         VARCHAR(50)     │   │   │ updated_at      TIMESTAMP       │   │
└─────────────────────────────────┘   │   └─────────────────────────────────┘   │
                                      │                                         │
┌─────────────────────────────────┐   │   ┌─────────────────────────────────┐   │
│          integrations           │   │   │          messages               │   │
├─────────────────────────────────┤   │   ├─────────────────────────────────┤   │
│ id (PK)         VARCHAR(36)     │   │   │ id (PK)         VARCHAR(36)     │   │
│ user_id (FK)    VARCHAR(36)     ├───┘   │ conversation_id VARCHAR(36)     ├───┘
│ channel         VARCHAR(50)     │       │ sender_type     ENUM            │
│ status          ENUM            │       │ content         TEXT            │
│ config          JSON            │       │ is_read         BOOLEAN         │
└─────────────────────────────────┘       └─────────────────────────────────┘
```

---

## ✨ Premium Accomplishments & Enhancements

During this upgrade phase, we implemented outstanding visual and structural adjustments:

### 1. dynamic Favicon Package
* **Generator Script**: Designed a custom Python script using `Pillow` to extract the high-resolution circle-and-three-dots JPG branding logo.
* **Compatibility Package**: Layer-packed a transparent `favicon.ico` (resolutions `16x16`, `32x32`, `48x48`, `64x64`), `favicon-32x32.png`, `favicon-16x16.png`, and a 180px `apple-touch-icon.png` in the public assets.
* **Linked Assembly**: Configured `index.html` header structure for robust loading across modern and legacy web browsers.

### 2. Sliding Network Banner System
* **Offline Banner Component**: Completely rewrote the connection visual indicator in `src/components/OfflineBanner.tsx`.
* **Transitions**: Leveraged custom `@keyframes slideDownBanner` and `slideUpBanner` in `src/index.css` to slide down smoothly when connection drops.
* **Auto-Fade**: When back online, a gorgeous green banner appears, flashes for 2.2 seconds, fades out over 300ms (`transition: opacity 0.3s`), and automatically unmounts, satisfying the 2.5-second online visual timer.

### 3. Polish Connected Channels Table
* **Connected Only Filter**: Refactored the dashboard's summary table to iterate exclusively on active integrations (`status === 'connected'`).
* **Interactive Empty State**: Designed a beautiful glass-morphic visual empty state containing floating illustrations and a CTA button prompting connection whenever no active channels exist.

### 4. Direct Overview Landing Page
* **Routing Redirection**: Updated the successful authentication block inside `login/page.tsx`.
* **Overview First**: Instead of directing to configuration wizard overlays, users landing on the platform are immediately welcomed by the visual metrics and charts of the **Overview** dashboard (`/`).

### 5. Multi-Modal Widget State Sync
* **Context State Hub**: Deployed a React context (`WidgetConfigContext`) wrapping the entire application router tree.
* **Widget-to-Channel Sync**: The widget customizer panel (**Web Widget** tab) and the channels connector popup (**Channels** tab) are bidirectionally wired into this context. 
* **Database Cohesion**: Toggling features, customizing bot names, positions, greetings, or changing brand color schemes updates both the React context states and backend `widget_configs` & `integrations` databases instantly—with absolutely zero page reloads.

---

## ⚙️ Installation & Local Deployment Guide

### Prerequisites
* Go 1.22 or higher
* Node.js v18 or higher (with npm)
* Python (with Pillow installed for favicon transformations)
* MySQL client or access to a TiDB Cloud database

---

### Step 1: Clone & Configure the Backend

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Create your `.env` configuration file:
   ```bash
   cp .env.example .env
   ```
3. Update `.env` with your active database connection strings and AI credentials:
   ```env
   PORT=8080
   DB_DSN="user:password@tcp(tidb-host:4000)/noant?tls=true&parseTime=true"
   REDIS_URL="redis://default:token@redis-host:6379"
   GROQ_API_KEY="gsk_..."
   JWT_SECRET="super-secure-key"
   ```
4. Run database migrations:
   ```bash
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/001_init.sql
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/005_user_isolation.sql
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/006_notifications_widget.sql
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/007_message_source.sql
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/008_inventory_leads.sql
   mysql -h your-tidb-host -P 4000 -u your-user -p noant < migrations/009_inventory_leads_fix.sql
   ```
5. Install Go dependencies and launch:
   ```bash
   go mod download
   go run main.go
   ```

---

### Step 2: Configure & Run the React Frontend

1. Open a new terminal and navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install npm dependencies:
   ```bash
   npm install
   ```
3. Generate high-quality assets using the Python script:
   ```bash
   python -c "
   from PIL import Image
   import os
   rgba = Image.open(r'../brain/d5e6425a-d5b3-48c4-b03c-36f16dbfa6fc/media__1779803513026.jpg').convert('RGBA')
   datas = rgba.getdata()
   newData = [(255,255,255,0) if item[0]>240 and item[1]>240 and item[2]>240 else item for item in datas]
   rgba.putdata(newData)
   sizes = {'logo.png': (192, 192), 'favicon-32x32.png': (32, 32), 'favicon-16x16.png': (16, 16), 'apple-touch-icon.png': (180, 180)}
   for name, size in sizes.items():
       rgba.resize(size, Image.Resampling.LANCZOS).save(os.path.join('public', name), 'PNG')
   rgba.save(os.path.join('public', 'favicon.ico'), format='ICO', sizes=[(16, 16), (32, 32), (48, 48)])
   print('Assets compiled!')
   "
   ```
4. Run the Vite development server:
   ```bash
   npm run dev
   ```

---

## 📑 Comprehensive API Directory

All backend endpoints are standardized with versioning prefix `/api/v1` and return JSON payloads.

| Category | Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- | :--- |
| **Authentication** | `POST` | `/auth/register` | Create a new tenant account | No |
| | `POST` | `/auth/login` | Log in and return JWT token | No |
| | `POST` | `/auth/refresh` | Refresh an expiring session token | Yes |
| | `POST` | `/auth/change-password` | Change user password | Yes |
| **Chats & Inbox** | `GET` | `/chats/conversations` | List conversations (paginated) | Yes |
| | `GET` | `/chats/conversations/:id` | Fetch specific conversation thread | Yes |
| | `POST` | `/chats/conversations/:id/messages` | Send an outbound message | Yes |
| | `PUT` | `/chats/conversations/:id/takeover` | Human takes over control from AI | Yes |
| | `POST` | `/chats/direct-chat` | Start interactive test with AI | Yes |
| **Training Engine**| `GET` | `/training/categories` | List Q&A categories | Yes |
| | `POST` | `/training/categories` | Create Q&A category | Yes |
| | `POST` | `/training/bulk-qa` | Bulk import trained QA pairs | Yes |
| | `POST` | `/training/csv-upload` | Upload and auto-parse CSV | Yes |
| | `GET` | `/training/unknown-questions` | List logged knowledge gaps | Yes |
| | `POST` | `/training/unknown-questions/:id/train` | Train and resolve unknown question | Yes |
| **Integrations** | `GET` | `/integrations/list` | Fetch active/inactive channel list | Yes |
| | `POST` | `/integrations/connect` | Connect channel configurations | Yes |
| | `POST` | `/integrations/disconnect/:channel`| Disconnect an integration channel | Yes |
| **Widget Config**  | `GET` | `/widget/config` | Load custom web widget config | Yes |
| | `POST` | `/widget/config` | Update custom web widget settings | Yes |
| **Analytics** | `GET` | `/analytics/overview` | Fetch primary stats and trends | Yes |
| | `GET` | `/analytics/insights` | Fetch custom AI business insights | Yes |

---

## 🛡️ License & Development Credits

* **Engineered by**: Advanced Agentic Coding Team — Google DeepMind.
* **Designed for**: Dynamic local and international enterprise systems.
* **License**: MIT License. Built with ❤️ for African businesses.
