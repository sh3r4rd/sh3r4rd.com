# sh3r4rd.com

This is my personal website that I use to demonstrate professional abilities throughout the tech stack. Anyone interested will also be able to send me a 
request for my resume.

## Tools
It is build using React, Tailwind and Vite on the front-end and uses AWS services like S3, Cloudfront, API Gateway and Lambda on the backend.

## Upcoming
I'm working on a new feature to leverage AI to compare my experience with prospective job descriptions to determine if I'm a good fit for a particular role.

## Deployment

### GitHub Actions (Recommended)

The project includes a GitHub Actions workflow that automatically deploys to S3 when code is pushed to the main branch.

**Required GitHub Secrets:**
- `AWS_ACCESS_KEY_ID` - AWS access key with S3 permissions
- `AWS_SECRET_ACCESS_KEY` - AWS secret access key
- `S3_BUCKET_NAME` - Name of your S3 bucket

To set up secrets:
1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add the three required secrets listed above

### Manual Deployment

Run the following command to deploy assets to production manually:

```bash
make deploy bucket=bucket-name
```

This requires AWS CLI to be configured locally.
