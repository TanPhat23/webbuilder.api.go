# Cloudinary Image Upload Integration

This document provides comprehensive information about the Cloudinary image upload integration in the WebBuilder API.

## Table of Contents

1. [Overview](#overview)
2. [Setup Instructions](#setup-instructions)
3. [Environment Variables](#environment-variables)
4. [API Endpoints](#api-endpoints)
5. [Usage Examples](#usage-examples)
6. [Error Handling](#error-handling)
7. [Security Considerations](#security-considerations)

## Overview

The WebBuilder API integrates with Cloudinary for robust, scalable image upload and management. This integration provides:

- **Multipart file upload** support for standard form-based uploads
- **Base64 image upload** for inline image data
- **Automatic image optimization** through Cloudinary's CDN
- **Secure storage** with user-specific folders
- **Soft delete** functionality for image management
- **Image metadata** tracking in PostgreSQL database

## Setup Instructions

### 1. Create a Cloudinary Account

1. Visit [Cloudinary](https://cloudinary.com/) and sign up for a free account
2. Navigate to your Dashboard to find your credentials
3. Note down the following values:
   - Cloud Name
   - API Key
   - API Secret

### 2. Configure Environment Variables

Create a `.env` file in the root directory (or update your existing one):

```env
CLOUDINARY_CLOUD_NAME=your_cloudinary_cloud_name
CLOUDINARY_API_KEY=your_cloudinary_api_key
CLOUDINARY_API_SECRET=your_cloudinary_api_secret
```

### 3. Database Migration

Ensure your PostgreSQL database has the `Image` table. The schema should match:

```sql
CREATE TABLE "Image" (
  "ImageId" VARCHAR PRIMARY KEY,
  "ImageLink" VARCHAR NOT NULL DEFAULT '',
  "ImageName" VARCHAR,
  "UserId" VARCHAR NOT NULL,
  "CreatedAt" TIMESTAMP(6) NOT NULL,
  "DeletedAt" TIMESTAMP(6),
  "UpdatedAt" TIMESTAMP(6) NOT NULL,
  FOREIGN KEY ("UserId") REFERENCES "User"("Id") ON DELETE CASCADE
);

CREATE INDEX "IX_Images_UserId" ON "Image"("UserId");
```

### 4. Install Dependencies

```bash
go mod download
```

## Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `CLOUDINARY_CLOUD_NAME` | Yes | Your Cloudinary cloud name | `my-cloud-name` |
| `CLOUDINARY_API_KEY` | Yes | Your Cloudinary API key | `123456789012345` |
| `CLOUDINARY_API_SECRET` | Yes | Your Cloudinary API secret | `abcdefghijklmnopqrstuvwxyz` |

## API Endpoints

All endpoints require authentication via the `AuthenticateMiddleware`.

### 1. Upload Image (Multipart)

**Endpoint:** `POST /api/v1/images`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `image` (required): The image file
- `imageName` (optional): Custom name for the image

**Response:** `201 Created`
```json
{
  "imageId": "clxyz123456789",
  "imageLink": "https://res.cloudinary.com/your-cloud/image/upload/v1234567890/webbuilder/user-id/image_name.jpg",
  "imageName": "My Image",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### 2. Upload Image (Base64)

**Endpoint:** `POST /api/v1/images/base64`

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "imageData": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "imageName": "My Image"
}
```

**Response:** `201 Created`
```json
{
  "imageId": "clxyz123456789",
  "imageLink": "https://res.cloudinary.com/your-cloud/image/upload/v1234567890/webbuilder/user-id/image.png",
  "imageName": "My Image",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### 3. Get All User Images

**Endpoint:** `GET /api/v1/images`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
[
  {
    "imageId": "clxyz123456789",
    "imageLink": "https://res.cloudinary.com/...",
    "imageName": "My Image",
    "userId": "user_abc123",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z"
  }
]
```

### 4. Get Image By ID

**Endpoint:** `GET /api/v1/images/:imageid`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "imageId": "clxyz123456789",
  "imageLink": "https://res.cloudinary.com/...",
  "imageName": "My Image",
  "userId": "user_abc123",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

### 5. Delete Image

**Endpoint:** `DELETE /api/v1/images/:imageid`

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `204 No Content`

**Note:** This performs a soft delete, setting the `DeletedAt` timestamp.

## Usage Examples

### cURL Examples

#### Upload an Image (Multipart)

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@/path/to/image.jpg" \
  -F "imageName=My Awesome Image"
```

#### Upload an Image (Base64)

```bash
curl -X POST http://localhost:8080/api/v1/images/base64 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "imageData": "data:image/png;base64,iVBORw0KGgo...",
    "imageName": "My Image"
  }'
```

#### Get All Images

```bash
curl -X GET http://localhost:8080/api/v1/images \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Delete an Image

```bash
curl -X DELETE http://localhost:8080/api/v1/images/clxyz123456789 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### JavaScript/Frontend Examples

#### Upload with FormData

```javascript
const uploadImage = async (file) => {
  const formData = new FormData();
  formData.append('image', file);
  formData.append('imageName', 'My Image');

  const response = await fetch('http://localhost:8080/api/v1/images', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });

  return await response.json();
};
```

#### Upload Base64 Image

```javascript
const uploadBase64Image = async (base64Data) => {
  const response = await fetch('http://localhost:8080/api/v1/images/base64', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      imageData: base64Data,
      imageName: 'My Image'
    })
  });

  return await response.json();
};
```

#### React Component Example

```jsx
import { useState } from 'react';

function ImageUploader() {
  const [uploading, setUploading] = useState(false);
  const [imageUrl, setImageUrl] = useState(null);

  const handleUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    setUploading(true);
    const formData = new FormData();
    formData.append('image', file);
    formData.append('imageName', file.name);

    try {
      const response = await fetch('/api/v1/images', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });

      const data = await response.json();
      setImageUrl(data.imageLink);
    } catch (error) {
      console.error('Upload failed:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input 
        type="file" 
        accept="image/*" 
        onChange={handleUpload}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
      {imageUrl && <img src={imageUrl} alt="Uploaded" />}
    </div>
  );
}
```

## Error Handling

### Common Error Responses

#### 400 Bad Request - Invalid File Type

```json
{
  "error": "Invalid image file",
  "errorMessage": "invalid file type: .pdf. Allowed types: jpg, jpeg, png, gif, webp, svg, bmp"
}
```

#### 400 Bad Request - File Too Large

```json
{
  "error": "Invalid image file",
  "errorMessage": "file size exceeds maximum allowed size of 10MB"
}
```

#### 401 Unauthorized

```json
{
  "error": "Unauthorized",
  "errorMessage": "You must be logged in to upload images"
}
```

#### 404 Not Found

```json
{
  "error": "Image not found"
}
```

#### 500 Internal Server Error

```json
{
  "error": "Failed to upload image to Cloudinary",
  "errorMessage": "cloudinary error details..."
}
```

## Security Considerations

### 1. File Validation

- **File size limit:** Maximum 10MB per file
- **Allowed formats:** jpg, jpeg, png, gif, webp, svg, bmp
- **MIME type checking:** Implemented at the file header level

### 2. User Isolation

- Images are stored in user-specific folders: `webbuilder/{userId}/`
- Users can only access and delete their own images
- Database queries always filter by `userId`

### 3. Authentication

- All endpoints require valid JWT authentication
- User ID is extracted from the authentication token
- Middleware validates tokens before processing requests

### 4. Soft Deletes

- Images are soft-deleted (not permanently removed)
- `DeletedAt` timestamp is set instead of actual deletion
- Allows for data recovery and audit trails

### 5. Cloudinary Security

- API credentials are stored in environment variables
- Secure HTTPS URLs are used for all image deliveries
- Cloudinary handles CDN security and DDoS protection

## Best Practices

1. **Image Optimization:** Cloudinary automatically optimizes images. Use URL transformations for responsive images:
   ```
   https://res.cloudinary.com/.../w_400,h_300,c_fill/image.jpg
   ```

2. **Error Handling:** Always implement proper error handling in your frontend:
   ```javascript
   try {
     const result = await uploadImage(file);
   } catch (error) {
     // Handle upload failure
   }
   ```

3. **Loading States:** Show loading indicators during uploads for better UX

4. **File Size Checking:** Validate file sizes client-side before uploading:
   ```javascript
   if (file.size > 10 * 1024 * 1024) {
     alert('File too large!');
     return;
   }
   ```

5. **Image Previews:** Show thumbnails before uploading using FileReader API

6. **Retry Logic:** Implement retry logic for failed uploads due to network issues

## Troubleshooting

### Issue: "Cloudinary credentials not found"

**Solution:** Ensure all three environment variables are set:
- `CLOUDINARY_CLOUD_NAME`
- `CLOUDINARY_API_KEY`
- `CLOUDINARY_API_SECRET`

### Issue: Upload fails with timeout

**Solution:** 
- Check your network connection
- Increase the context timeout in the handler (default: 30 seconds)
- Verify Cloudinary service status

### Issue: Images not appearing after upload

**Solution:**
- Verify the image URL is valid
- Check CORS settings if accessing from a different domain
- Ensure the user has permission to view the image

### Issue: Database foreign key constraint error

**Solution:**
- Ensure the `User` table exists with the correct `Id` column
- Verify the authenticated user exists in the database
- Check the `UserId` foreign key relationship

## Additional Resources

- [Cloudinary Documentation](https://cloudinary.com/documentation)
- [Cloudinary Go SDK](https://github.com/cloudinary/cloudinary-go)
- [Image Transformation Reference](https://cloudinary.com/documentation/image_transformations)
- [Upload API Reference](https://cloudinary.com/documentation/image_upload_api_reference)

## Support

For issues or questions:
1. Check this documentation first
2. Review the error messages carefully
3. Verify environment variables are set correctly
4. Check Cloudinary dashboard for quota limits
5. Review application logs for detailed error information