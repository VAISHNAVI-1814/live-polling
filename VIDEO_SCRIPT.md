# Video Submission Script (3–5 Minutes)
### GUVI Developer Internship Developer Task — Live Polling Tool

Use this script and step-by-step cue sheet when recording your 3–5 minute walk-through video for submission to `devhiring@hclguvi.com`.

---

## ⏱️ Video Outline & Timing Breakdown

| Timestamp | Section | Key Visual / Action | Spoken Talking Points |
|---|---|---|---|
| **0:00 – 0:30** | **Introduction** | Show homepage / dashboard on screen. Webcam on. | Introduce yourself, state the project goal: a real-time live polling platform where audience votes update all screens with zero page refreshes using React, Go (Gin), MongoDB, and Redis. |
| **0:30 – 1:00** | **Authentication** | Sign up a new user or log in. | Explain that poll creation is protected with JWT authentication and bcrypt password hashing, ensuring polls can only be created and managed by authenticated organizers. |
| **1:00 – 1:40** | **Create Poll** | Navigate to `/create-poll`. Fill in question & options. | Demonstrate dynamic option addition (min 2, max 10), duration selector, backend input validation, and live preview. Click "Publish Live Poll". |
| **1:40 – 2:10** | **Share Poll** | Show the Live Results page. Click "Share". | Highlight the unique 6-character Share Code and audience link. Show how anyone with the link can join the poll without needing an account. |
| **2:10 – 3:00** | **Multi-Browser Live Vote Demo (Crucial)** | Place two browser windows side by side (Browser A: Results, Browser B: Voter). Cast vote in Browser B. | **Key Moment**: Show Browser A updating immediately—progress bars animate and numbers increment without touching the refresh button! Point to the pulsing green "Live Stream Active" indicator. |
| **3:00 – 3:40** | **Architecture Explanation** | Show the architecture diagram or codebase structure. | Explain the data flow: Client vote -> Go (Gin) validates -> Redis `HINCRBY` atomic counter + `SADD` duplicate check -> MongoDB durable persistence -> Redis Pub/Sub event -> Go WebSocket Hub -> instant broadcast to all browsers. Emphasize that Redis is genuinely doing the real-time heavy lifting. |
| **3:40 – 4:20** | **Biggest Challenge & How You Solved It** | Show `redis_service.go` code snippet. | Discuss race conditions and concurrency. Explain why traditional SQL/Mongo updates struggle under simultaneous votes and how you solved it using Redis atomic in-memory primitives (`HINCRBY` and `SADD`) with a dual-layer duplicate voter check. |
| **4:20 – 5:00** | **AI Usage & Conclusion** | Webcam / final closing screen. | Address the AI question openly and honestly: Explain how AI was used as a pair programming assistant to accelerate boilerplate and styling, while you mastered the distributed real-time flow, Redis concurrency, and WebSocket lifecycle to prepare for technical interviews. Conclude with thank you. |

---

## 🎙️ Word-for-Word Script Guide

### 0:00 – 0:30 | Introduction
> *"Hello everyone! My name is [Your Name], and this is my submission for the GUVI Developer Internship Developer Task: the Live Polling Application. The objective of this project is to create a seamless live polling platform where an organizer can create a poll, share a link, and collect audience votes with all results updating live in real-time—with absolutely no page refreshes required. To achieve this, I built the application using React on the frontend, Go with Gin on the backend, MongoDB for durable persistence, and Redis for high-speed atomic counts and real-time Pub/Sub message distribution."*

### 0:30 – 1:00 | Authentication & Security
> *"Let's begin with authentication. The project specification requires that poll creation is protected. Here on the login page, users can either register a new account or log in. Passwords are securely hashed using bcrypt on the Go backend, and successful authentication issues a signed JWT token. The dashboard route is protected by both React route guards and Go backend middleware. Once logged in, the organizer accesses their personal dashboard, showing active polls, total votes received, and quick management controls."*

### 1:00 – 1:40 | Poll Creation & Validation
> *"Now let's create a poll. Clicking 'Create New Poll' takes us to this form. We can enter our question—for example, 'What is your favorite cloud provider?' We can dynamically add or remove options with a minimum of 2 and maximum of 10. The backend validates that all options are non-empty, unique, and sanitized. We can also set an expiration duration—like 1 hour, 24 hours, or never. As we type, a live voter preview shows how it will appear to our audience. Let's click 'Publish Live Poll'."*

### 1:40 – 2:10 | Shareable Link & Code
> *"Immediately upon creation, we are redirected to the Live Results screen and given a unique 6-character share code and audience voting URL. Organizers can project this live results screen on an auditorium display or webinar stream while attendees scan or click the share link to vote from their own devices."*

### 2:10 – 3:00 | Multi-Browser Real-Time Test (The Core Requirement)
> *"Now for the most critical test: demonstrating true real-time synchronization. On the left side of my screen, I have Browser 1 showing the Live Results view. On the right, in an incognito window, I have opened the public voter link as an audience member.*
>
> *Notice the voter page displays our choices: AWS, Google Cloud, and Azure. In Browser 2, I select 'Google Cloud' and click 'Cast Live Vote'.*
>
> *Watch Browser 1 on the left closely: immediately, the count increments from 0 to 1, the progress bar smoothly expands, the percentage updates to 100%, and the leader badge appears—all without any page reload! The green indicator confirms an active full-duplex WebSocket connection. If another voter votes for AWS from a third browser, it updates instantly for everyone watching."*

### 3:00 – 3:40 | Architecture & Stack Walkthrough
> *"Here is how the architecture achieves this:*
> 1. *When an audience member clicks vote, an HTTP POST request is sent to the Go Gin backend.*
> 2. *Go validates the input and checks that the poll is active and not expired.*
> 3. *Next comes Redis: rather than doing slow database locking, Redis performs an atomic `HINCRBY` on the poll's vote hash and records the voter token in a Redis Set via `SADD` to prevent double-voting in O(1) time.*
> 4. *The vote is also recorded in MongoDB for durable audit persistence.*
> 5. *Finally, Redis publishes a `VOTE_UPDATED` event through its Pub/Sub channel. The Go WebSocket Hub listening on this channel receives the message and pushes it out across all open WebSocket connections in milliseconds."*

### 3:40 – 4:20 | Biggest Challenge & How I Solved It
> *"The biggest engineering challenge I encountered was preventing race conditions under high concurrent voting while ensuring that the duplicate-vote prevention didn't slow down the response time.*
>
> *If hundreds of people vote in the same second, traditional database queries create lock contention. I solved this by implementing a dual-layer strategy: Redis handles the hot path completely in-memory using `HINCRBY` and `SADD`, guaranteeing atomic serialized execution without race conditions. MongoDB is then updated safely. This decoupled design allows the app to handle high vote spikes effortlessly."*

### 4:20 – 5:00 | AI Usage & Final Conclusion
> *"Regarding AI tools: I used the Antigravity AI coding assistant to help scaffold initial repetitive structures like Docker configs and boilerplate model definitions, which allowed me to focus on the core real-time architecture, Redis atomic commands, and WebSocket concurrency.*
>
> *I made sure to dive deep into every line of code, understand the goroutine lifecycles, and thoroughly test the system so that I am fully prepared to discuss every technical decision during the interview rounds.*
>
> *Thank you for reviewing my project! You can find the public GitHub repository and live deployment links in the submission email to devhiring@hclguvi.com. Have a great day!"*
