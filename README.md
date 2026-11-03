# FindIt: Campus Lost & Found Portal

![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?logo=gin&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white)
![JWT](https://img.shields.io/badge/Auth-JWT-000000?logo=jsonwebtokens&logoColor=white)
![PWA](https://img.shields.io/badge/PWA-installable-5A0FC8?logo=pwa&logoColor=white)

FindIt is a web portal where students report lost items, post found items, and get belongings back to their real owners.

**Live demo:** [https://YOUR-APP.onrender.com](https://YOUR-APP.onrender.com)
*(The free server may take about 30 seconds to wake up on the first visit.)*

<p align="center">
  <img src="docs/screenshots/landing_home.jpg" alt="FindIt landing page" width="85%">
</p>

## The Problem

Students lose ID cards, bottles, chargers and notebooks every day. Messages in WhatsApp groups get buried within hours, and there is no search, no photo and no way to check who the real owner is. Many items never get back to their owners.

## The Solution

FindIt gives the whole campus one searchable place to report and find items. A person who thinks an item is theirs must **describe it**, and the poster **approves or rejects** the claim. Contact details stay hidden until a claim is approved.

## Features

- **Secure login:** register and log in with an email and password. Passwords are stored as bcrypt hashes, and logins use JWT tokens.
- **Post lost or found items:** name, category, place, description, contact and a photo.
- **Where found, where kept:** every found item shows both places, so the owner knows where to collect it.
- **Search and filters:** by keyword, category, location and type. Capital or small letters do not matter.
- **Claim and verification:** the claimer describes the item and the poster approves or rejects. Only one claim per person per item, and nobody can claim their own post.
- **Private contacts:** the contact is shown only to the poster and to the approved claimer.
- **Item status:** open, claimed or returned. Only the poster can edit, delete or mark an item as returned.
- **Photos in the database:** photos are saved in PostgreSQL, so they stay safe when the server restarts.
- **Live counters:** the landing page shows how many items are found, lost and returned.
- **Installable app (PWA):** install FindIt on a phone or laptop and open it like a normal app.

## Screenshots

| Browse items | Post an item |
|---|---|
| <img src="docs/screenshots/browse.jpg" alt="Browse items"> | <img src="docs/screenshots/post_form.jpg" alt="Post an item"> |

| Claim form | Claim approval |
|---|---|
| <img src="docs/screenshots/claim_form.jpg" alt="Claim form"> | <img src="docs/screenshots/claims_review.jpg" alt="Poster reviews a claim"> |

## Tech Stack

| Part | Technology |
|---|---|
| Backend | Go, Gin |
| Database | PostgreSQL, GORM |
| Security | JWT, bcrypt |
| Frontend | HTML, CSS, JavaScript |
| App features | Web app manifest, service worker (PWA) |
| Hosting | Render |

## How It Works

```
Browser  ->  Go server (Gin + JWT check)  ->  PostgreSQL (users, items + photos, claims)
```

1. A student registers, logs in, and gets a token.
2. The student posts a lost or found item with a photo.
3. Others search the items and click **This is mine** on an item.
4. The poster reads the description and approves or rejects the claim.
5. When approved, the contact is shared. The poster marks the item as **returned**.

## Project Structure

```
FindIt/
├── main.go             # starts the server and lists all routes
├── auth/               # JWT creation and login checks
├── database/           # PostgreSQL connection and tables
├── models/             # User, Item and Claim
├── handlers/           # register/login, items, claims
└── web/                # website, manifest.json, sw.js and icons
```

## Run It on Your Computer

**You need:** [Go](https://go.dev/dl/) and [PostgreSQL](https://www.postgresql.org/download/).

```bash
git clone https://github.com/Gouri-Pawar/FindIt.git
cd FindIt

# create the database
createdb lostfound

# start the server
go run main.go
```

Open **http://localhost:8080** in your browser. The tables are created automatically when the server starts.

### Settings (optional)

| Variable | Meaning | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | local database named `lostfound` |
| `JWT_SECRET` | Secret used to sign login tokens | a development value (change it online) |

Example:

```bash
export DATABASE_URL="host=localhost user=postgres password=secret dbname=lostfound port=5432 sslmode=disable"
export JWT_SECRET="use-a-long-random-text"
go run main.go
```

## Main API Routes

| Method | Route | Who can use it | What it does |
|---|---|---|---|
| POST | `/register` | Anyone | Create an account |
| POST | `/login` | Anyone | Log in and get a token |
| GET | `/items` | Anyone | List and search items |
| GET | `/items/:id` | Anyone | Get one item |
| GET | `/items/:id/image` | Anyone | Get the photo of an item |
| POST | `/items` | Logged in | Post an item |
| PUT | `/items/:id` | Poster only | Edit an item or mark it returned |
| DELETE | `/items/:id` | Poster only | Delete an item |
| POST | `/items/:id/claims` | Logged in | Send a claim |
| GET | `/items/:id/claims` | Poster only | See claims on my item |
| GET | `/claims/mine` | Logged in | See claims I sent |
| PUT | `/claims/:id` | Poster only | Approve or reject a claim |

Search example:

```
GET /items?type=found&category=bottle&location=library&q=steel&page=1
```

## Deploy on Render

1. Create a **PostgreSQL** database on Render.
2. Create a **Web Service** from this GitHub repository.
3. Add the environment variables `DATABASE_URL` and `JWT_SECRET`.
4. Deploy. Every push to GitHub deploys the new version.

## Roadmap

- [ ] Smart match suggestions between lost and found items
- [ ] Email or push notifications for claims
- [ ] Instant page refresh with live updates (the server side is ready)
- [ ] Admin panel to remove spam posts
- [ ] My Posts page with editing on the website
- [ ] Support for many colleges

## Author

**Gouri Pawar**
[GitHub](https://github.com/Gouri-Pawar) · [LinkedIn](https://www.linkedin.com/in/gouri-pawar-a55537271)

If you like this project, please give it a star.
