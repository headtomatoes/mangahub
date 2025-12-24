# Rating and Comment API Testing Guide

This document provides detailed JSON test cases for testing the Rating and Comment APIs.

## Table of Contents
- [Authentication Setup](#authentication-setup)
- [Rating API Tests](#rating-api-tests)
- [Comment API Tests](#comment-api-tests)
- [Error Scenarios](#error-scenarios)

---

## Authentication Setup

First, you need to authenticate to get a JWT token:

### 1. Register a User
```bash
POST http://localhost:8084/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "testuser@example.com",
  "password": "Password123!"
}
```

**Expected Response:**
```json
{
  "message": "user registered successfully",
  "user": {
    "id": "uuid-here",
    "username": "testuser",
    "email": "testuser@example.com",
    "role": "user"
  }
}
```

### 2. Login to Get Token
```bash
POST http://localhost:8084/auth/login
Content-Type: application/json

{
  "email": "testuser@example.com",
  "password": "Password123!"
}
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Save the `access_token` and use it in the Authorization header for all protected endpoints:**
```
Authorization: Bearer <access_token>
```

---

## Rating API Tests

### 1. Create or Update Rating

**Endpoint:** `POST /api/manga/:manga_id/ratings`  
**Authentication:** Required  
**Description:** Creates a new rating or updates an existing rating for the authenticated user.

#### Test Case 1: Create First Rating
```bash
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 8
}
```

**Expected Response (201 OK):**
```json
{
  "id": 1,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "rating": 8,
  "created_at": "2025-11-23T10:30:00Z",
  "updated_at": "2025-11-23T10:30:00Z"
}
```

#### Test Case 2: Update Existing Rating
```bash
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 9
}
```

**Expected Response (200 OK):**
```json
{
  "id": 1,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "rating": 9,
  "created_at": "2025-11-23T10:30:00Z",
  "updated_at": "2025-11-23T10:35:00Z"
}
```

#### Test Case 3: Invalid Rating (Too Low)
```bash
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 0
}
```

**Expected Response (400 Bad Request):**
```json
{
  "error": "Key: 'CreateRatingDTO.Rating' Error:Field validation for 'Rating' failed on the 'min' tag"
}
```

#### Test Case 4: Invalid Rating (Too High)
```bash
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 11
}
```

**Expected Response (400 Bad Request):**
```json
{
  "error": "Key: 'CreateRatingDTO.Rating' Error:Field validation for 'Rating' failed on the 'max' tag"
}
```

---

### 2. Get User's Own Rating

**Endpoint:** `GET /api/manga/:manga_id/ratings/me`  
**Authentication:** Required  
**Description:** Retrieves the authenticated user's rating for a specific manga.

```bash
GET http://localhost:8084/api/manga/1/ratings/me
Authorization: Bearer <access_token>
```

**Expected Response (200 OK):**
```json
{
  "rating": 9,
  "created_at": "2025-11-23T10:30:00Z",
  "updated_at": "2025-11-23T10:35:00Z"
}
```

**If No Rating Exists (404 Not Found):**
```json
{
  "error": "rating not found"
}
```

---

### 3. Get All Ratings for a Manga

**Endpoint:** `GET /api/manga/:manga_id/ratings?page=1&page_size=20`  
**Authentication:** Not Required (Public)  
**Description:** Retrieves all ratings for a manga with pagination.

```bash
GET http://localhost:8084/api/manga/1/ratings?page=1&page_size=10
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": 5,
      "user_id": "uuid-5",
      "username": "alice",
      "manga_id": 1,
      "rating": 10,
      "created_at": "2025-11-23T12:00:00Z",
      "updated_at": "2025-11-23T12:00:00Z"
    },
    {
      "id": 3,
      "user_id": "uuid-3",
      "username": "bob",
      "manga_id": 1,
      "rating": 8,
      "created_at": "2025-11-23T11:30:00Z",
      "updated_at": "2025-11-23T11:30:00Z"
    },
    {
      "id": 1,
      "user_id": "uuid-1",
      "username": "testuser",
      "manga_id": 1,
      "rating": 9,
      "created_at": "2025-11-23T10:30:00Z",
      "updated_at": "2025-11-23T10:35:00Z"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 3,
  "total_pages": 1
}
```

---

### 4. Get Average Rating for a Manga

**Endpoint:** `GET /api/manga/:manga_id/ratings/average`  
**Authentication:** Not Required (Public)  
**Description:** Retrieves the average rating and total count for a manga.

```bash
GET http://localhost:8084/api/manga/1/ratings/average
```

**Expected Response (200 OK):**
```json
{
  "average_rating": 9.0,
  "total_ratings": 3
}
```

---

### 5. Delete Rating

**Endpoint:** `DELETE /api/manga/:manga_id/ratings`  
**Authentication:** Required  
**Description:** Deletes the authenticated user's rating for a manga.

```bash
DELETE http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <access_token>
```

**Expected Response (200 OK):**
```json
{
  "message": "Rating deleted successfully"
}
```

**If No Rating Exists (404 Not Found):**
```json
{
  "error": "rating not found"
}
```

---

## Comment API Tests

### 1. Create Comment

**Endpoint:** `POST /api/manga/:manga_id/comments`  
**Authentication:** Required  
**Description:** Creates a new comment on a manga.

#### Test Case 1: Valid Comment
```bash
POST http://localhost:8084/api/manga/1/comments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "This manga is absolutely amazing! The art style is beautiful and the story keeps me hooked."
}
```

**Expected Response (201 Created):**
```json
{
  "id": 1,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "content": "This manga is absolutely amazing! The art style is beautiful and the story keeps me hooked.",
  "created_at": "2025-11-23T10:40:00Z",
  "updated_at": "2025-11-23T10:40:00Z"
}
```

#### Test Case 2: Empty Comment (Invalid)
```bash
POST http://localhost:8084/api/manga/1/comments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": ""
}
```

**Expected Response (400 Bad Request):**
```json
{
  "error": "Key: 'CreateCommentDTO.Content' Error:Field validation for 'Content' failed on the 'min' tag"
}
```

#### Test Case 3: Long Comment
```bash
POST http://localhost:8084/api/manga/1/comments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "This is a very detailed review of the manga. I've been reading it for months now and I must say that the character development is exceptional. The main protagonist has grown so much throughout the series, and the supporting cast each have their own unique personalities and backstories that make them feel real and relatable. The plot twists are unexpected yet logical when you look back at the foreshadowing. The pacing is perfect - not too fast, not too slow. Overall, this is a masterpiece that deserves all the recognition it gets."
}
```

**Expected Response (201 Created):**
```json
{
  "id": 2,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "content": "This is a very detailed review of the manga. I've been reading it for months now and I must say that the character development is exceptional...",
  "created_at": "2025-11-23T10:45:00Z",
  "updated_at": "2025-11-23T10:45:00Z"
}
```

---

### 2. Get All Comments for a Manga

**Endpoint:** `GET /api/manga/:manga_id/comments?page=1&page_size=20`  
**Authentication:** Not Required (Public)  
**Description:** Retrieves all comments for a manga with pagination.

```bash
GET http://localhost:8084/api/manga/1/comments?page=1&page_size=10
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": 2,
      "user_id": "user-uuid-here",
      "username": "testuser",
      "manga_id": 1,
      "content": "This is a very detailed review of the manga. I've been reading it for months now...",
      "created_at": "2025-11-23T10:45:00Z",
      "updated_at": "2025-11-23T10:45:00Z"
    },
    {
      "id": 1,
      "user_id": "user-uuid-here",
      "username": "testuser",
      "manga_id": 1,
      "content": "This manga is absolutely amazing! The art style is beautiful and the story keeps me hooked.",
      "created_at": "2025-11-23T10:40:00Z",
      "updated_at": "2025-11-23T10:40:00Z"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 2,
  "total_pages": 1
}
```

---

### 3. Get Specific Comment by ID

**Endpoint:** `GET /api/comments/:id`  
**Authentication:** Required  
**Description:** Retrieves a specific comment by its ID.

```bash
GET http://localhost:8084/api/comments/1
Authorization: Bearer <access_token>
```

**Expected Response (200 OK):**
```json
{
  "id": 1,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "content": "This manga is absolutely amazing! The art style is beautiful and the story keeps me hooked.",
  "created_at": "2025-11-23T10:40:00Z",
  "updated_at": "2025-11-23T10:40:00Z"
}
```

**If Comment Not Found (404 Not Found):**
```json
{
  "error": "comment not found"
}
```

---

### 4. Update Comment

**Endpoint:** `PUT /api/comments/:id`  
**Authentication:** Required  
**Description:** Updates a comment (only the owner can update).

#### Test Case 1: Valid Update
```bash
PUT http://localhost:8084/api/comments/1
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "Updated: This manga is still amazing! I just finished the latest chapter and it was incredible."
}
```

**Expected Response (200 OK):**
```json
{
  "id": 1,
  "user_id": "user-uuid-here",
  "username": "testuser",
  "manga_id": 1,
  "content": "Updated: This manga is still amazing! I just finished the latest chapter and it was incredible.",
  "created_at": "2025-11-23T10:40:00Z",
  "updated_at": "2025-11-23T11:00:00Z"
}
```

#### Test Case 2: Update Someone Else's Comment
```bash
PUT http://localhost:8084/api/comments/999
Authorization: Bearer <different_user_token>
Content-Type: application/json

{
  "content": "Trying to update someone else's comment"
}
```

**Expected Response (403 Forbidden):**
```json
{
  "error": "you don't have permission to update this comment"
}
```

---

### 5. Get Current User's Comments

**Endpoint:** `GET /api/comments/me?page=1&page_size=20`  
**Authentication:** Required  
**Description:** Retrieves all comments made by the authenticated user.

```bash
GET http://localhost:8084/api/comments/me?page=1&page_size=10
Authorization: Bearer <access_token>
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": 5,
      "user_id": "user-uuid-here",
      "username": "testuser",
      "manga_id": 3,
      "content": "Another great manga!",
      "created_at": "2025-11-23T12:00:00Z",
      "updated_at": "2025-11-23T12:00:00Z"
    },
    {
      "id": 2,
      "user_id": "user-uuid-here",
      "username": "testuser",
      "manga_id": 1,
      "content": "This is a very detailed review of the manga...",
      "created_at": "2025-11-23T10:45:00Z",
      "updated_at": "2025-11-23T10:45:00Z"
    },
    {
      "id": 1,
      "user_id": "user-uuid-here",
      "username": "testuser",
      "manga_id": 1,
      "content": "Updated: This manga is still amazing!...",
      "created_at": "2025-11-23T10:40:00Z",
      "updated_at": "2025-11-23T11:00:00Z"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 3,
  "total_pages": 1
}
```

---

### 6. Delete Comment

**Endpoint:** `DELETE /api/comments/:id`  
**Authentication:** Required  
**Description:** Deletes a comment (only the owner can delete).

```bash
DELETE http://localhost:8084/api/comments/1
Authorization: Bearer <access_token>
```

**Expected Response (200 OK):**
```json
{
  "message": "Comment deleted successfully"
}
```

**If Comment Not Found or Not Owner (404 Not Found):**
```json
{
  "error": "comment not found or you don't have permission to delete it"
}
```

---

## Error Scenarios

### 1. Missing Authorization Header

```bash
POST http://localhost:8084/api/manga/1/ratings
Content-Type: application/json

{
  "rating": 8
}
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "missing authorization header"
}
```

---

### 2. Invalid Token

```bash
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer invalid_token_here
Content-Type: application/json

{
  "rating": 8
}
```

**Expected Response (401 Unauthorized):**
```json
{
  "error": "invalid token"
}
```

---

### 3. Non-existent Manga

```bash
POST http://localhost:8084/api/manga/99999/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 8
}
```

**Expected Response (404 Not Found):**
```json
{
  "error": "manga not found"
}
```

---

### 4. Invalid Manga ID

```bash
POST http://localhost:8084/api/manga/invalid/ratings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "rating": 8
}
```

**Expected Response (400 Bad Request):**
```json
{
  "error": "Invalid manga ID"
}
```

---

## Complete Testing Flow

### Step 1: Setup Test User
```bash
# Register
POST http://localhost:8084/auth/register
Content-Type: application/json

{
  "username": "tester1",
  "email": "tester1@example.com",
  "password": "Test123!"
}

# Login and save token
POST http://localhost:8084/auth/login
Content-Type: application/json

{
  "email": "tester1@example.com",
  "password": "Test123!"
}
```

### Step 2: Test Rating Flow
```bash
# 1. Create rating
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <token>
Content-Type: application/json
{"rating": 8}

# 2. Get your rating
GET http://localhost:8084/api/manga/1/ratings/me
Authorization: Bearer <token>

# 3. Update rating
POST http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <token>
Content-Type: application/json
{"rating": 9}

# 4. Check average rating
GET http://localhost:8084/api/manga/1/ratings/average

# 5. Get all ratings
GET http://localhost:8084/api/manga/1/ratings?page=1&page_size=10

# 6. Delete rating
DELETE http://localhost:8084/api/manga/1/ratings
Authorization: Bearer <token>
```

### Step 3: Test Comment Flow
```bash
# 1. Create comment
POST http://localhost:8084/api/manga/1/comments
Authorization: Bearer <token>
Content-Type: application/json
{"content": "Great manga!"}

# 2. Get manga comments
GET http://localhost:8084/api/manga/1/comments?page=1&page_size=10

# 3. Get your comments
GET http://localhost:8084/api/comments/me?page=1&page_size=10
Authorization: Bearer <token>

# 4. Update comment (use comment ID from step 1)
PUT http://localhost:8084/api/comments/1
Authorization: Bearer <token>
Content-Type: application/json
{"content": "Updated: Even better than I thought!"}

# 5. Get specific comment
GET http://localhost:8084/api/comments/1
Authorization: Bearer <token>

# 6. Delete comment
DELETE http://localhost:8084/api/comments/1
Authorization: Bearer <token>
```

---

## Using cURL Commands

### Rating Examples

```bash
# Create/Update Rating
curl -X POST http://localhost:8084/api/manga/1/ratings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"rating": 8}'

# Get User's Rating
curl -X GET http://localhost:8084/api/manga/1/ratings/me \
  -H "Authorization: Bearer <token>"

# Get All Ratings
curl -X GET "http://localhost:8084/api/manga/1/ratings?page=1&page_size=10"

# Get Average Rating
curl -X GET http://localhost:8084/api/manga/1/ratings/average

# Delete Rating
curl -X DELETE http://localhost:8084/api/manga/1/ratings \
  -H "Authorization: Bearer <token>"
```

### Comment Examples

```bash
# Create Comment
curl -X POST http://localhost:8084/api/manga/1/comments \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "This manga is amazing!"}'

# Get Manga Comments
curl -X GET "http://localhost:8084/api/manga/1/comments?page=1&page_size=10"

# Get My Comments
curl -X GET "http://localhost:8084/api/comments/me?page=1&page_size=10" \
  -H "Authorization: Bearer <token>"

# Update Comment
curl -X PUT http://localhost:8084/api/comments/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "Updated comment text"}'

# Delete Comment
curl -X DELETE http://localhost:8084/api/comments/1 \
  -H "Authorization: Bearer <token>"
```

---

## Notes

1. **Pagination**: Default page size is 20, maximum is 100
2. **Rating Range**: 1-10 (inclusive)
3. **Comment Length**: Minimum 1 character, maximum 5000 characters
4. **Timestamps**: All timestamps are in UTC with RFC3339 format
5. **Authentication**: Use the token from the login response in the Authorization header
6. **Auto-Migration**: Rating and Comment tables are auto-created on server startup
7. **Average Rating**: Automatically recalculated whenever a rating is created, updated, or deleted

---

## Testing with Postman

1. Create a new collection called "Manga Rating & Comments"
2. Set up an environment variable `base_url` = `http://localhost:8084`
3. Set up an environment variable `token` = (paste your access token here)
4. Use `{{base_url}}` and `{{token}}` in your requests
5. Import the examples above into separate requests
6. Use the Collection Runner to run all tests in sequence
