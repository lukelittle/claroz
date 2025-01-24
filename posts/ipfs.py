import os
import requests
from django.conf import settings
from urllib.parse import urljoin

class IPFSClient:
    """Client for interacting with IPFS."""
    
    def __init__(self, api_url=None):
        self.api_url = api_url or os.getenv('IPFS_API_URL', 'http://localhost:5001/api/v0')
        self.gateway_url = os.getenv('IPFS_GATEWAY_URL', 'http://localhost:8080/ipfs')

    def add_file(self, file_path):
        """Upload a file to IPFS and return its CID."""
        try:
            with open(file_path, 'rb') as f:
                files = {'file': f}
                response = requests.post(urljoin(self.api_url, 'add'), files=files)
                response.raise_for_status()
                return response.json()['Hash']
        except Exception as e:
            raise Exception(f"Failed to upload file to IPFS: {str(e)}")

    def add_bytes(self, content):
        """Upload bytes to IPFS and return its CID."""
        try:
            files = {'file': content}
            response = requests.post(urljoin(self.api_url, 'add'), files=files)
            response.raise_for_status()
            return response.json()['Hash']
        except Exception as e:
            raise Exception(f"Failed to upload content to IPFS: {str(e)}")

    def get_file(self, cid):
        """Get a file from IPFS by its CID."""
        try:
            response = requests.post(
                urljoin(self.api_url, 'cat'), 
                params={'arg': cid}
            )
            response.raise_for_status()
            return response.content
        except Exception as e:
            raise Exception(f"Failed to get file from IPFS: {str(e)}")

    def get_gateway_url(self, cid):
        """Get the gateway URL for a CID."""
        return f"{self.gateway_url}/{cid}"

class IPFSStorage:
    """Custom storage backend for Django that uses IPFS."""
    
    def __init__(self):
        self.client = IPFSClient()
        self.location = settings.MEDIA_ROOT
        self.base_url = settings.MEDIA_URL.rstrip('/')

    def _save(self, name, content):
        """Save a file to IPFS and return the CID."""
        # Create full path
        full_path = os.path.join(self.location, name)
        directory = os.path.dirname(full_path)
        
        # Ensure directory exists
        if not os.path.exists(directory):
            os.makedirs(directory)
            
        # Save file locally first (required for some Django functionality)
        with open(full_path, 'wb') as f:
            for chunk in content.chunks():
                f.write(chunk)
                
        # Upload to IPFS
        cid = self.client.add_file(full_path)
        
        # Return the CID as the "name"
        return cid

    def _open(self, cid, mode='rb'):
        """Open a file from IPFS."""
        content = self.client.get_file(cid)
        return content

    def url(self, cid):
        """Get the URL for a file."""
        return self.client.get_gateway_url(cid)

    def get_available_name(self, name, max_length=None):
        """
        Return a filename that's free on the target storage system.
        Since we're using IPFS with content-addressed storage,
        we don't need to worry about name collisions.
        """
        return name

    def get_valid_name(self, name):
        """
        Return a filename suitable for use with the underlying storage system.
        Since we're using IPFS with content-addressed storage,
        we don't need to modify the name.
        """
        return name

    def path(self, name):
        """
        Return a local filesystem path where the file can be retrieved.
        For IPFS, we'll return the full path in the media directory.
        """
        return os.path.join(self.location, name)

    def delete(self, name):
        """
        Delete the specified file from storage.
        For IPFS, we only delete the local copy since IPFS content is immutable.
        """
        full_path = self.path(name)
        if os.path.exists(full_path):
            os.remove(full_path)
