from django.http import JsonResponse
from django.views.decorators.http import require_http_methods
from django.views.decorators.csrf import csrf_exempt
from django.contrib.auth.decorators import login_required
from django.core.cache import cache
import json
import requests
from datetime import datetime, timedelta
from .models import Profile, Post
import logging

logger = logging.getLogger(__name__)

class ATProtocolClient:
    """Client for interacting with AT Protocol-compatible servers."""
    
    def __init__(self, server_url, access_token=None):
        self.server_url = server_url.rstrip('/')
        self.access_token = access_token
        self.session = requests.Session()
        if access_token:
            self.session.headers.update({'Authorization': f'Bearer {access_token}'})

    def _make_request(self, method, endpoint, **kwargs):
        """Make an HTTP request to the AT Protocol server."""
        url = f"{self.server_url}/{endpoint.lstrip('/')}"
        try:
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            logger.error(f"AT Protocol request failed: {str(e)}")
            raise

    def get_profile(self, handle):
        """Fetch a profile from the federation server."""
        cache_key = f'federation_profile_{handle}'
        cached_profile = cache.get(cache_key)
        if cached_profile:
            return cached_profile

        try:
            data = self._make_request('GET', f'/xrpc/app.bsky.actor.getProfile', params={'actor': handle})
            profile_data = {
                'did': data.get('did'),
                'handle': data.get('handle'),
                'display_name': data.get('displayName'),
                'description': data.get('description'),
                'avatar': data.get('avatar'),
                'followers_count': data.get('followersCount', 0),
                'follows_count': data.get('followsCount', 0),
                'posts_count': data.get('postsCount', 0),
            }
            # Cache profile for 1 hour
            cache.set(cache_key, profile_data, 3600)
            return profile_data
        except Exception as e:
            logger.error(f"Failed to fetch profile {handle}: {str(e)}")
            return None

    def get_posts(self, handle=None, cursor=None, limit=20):
        """Fetch posts from the federation server."""
        try:
            params = {'limit': limit}
            if handle:
                params['actor'] = handle
            if cursor:
                params['cursor'] = cursor

            data = self._make_request('GET', '/xrpc/app.bsky.feed.getAuthorFeed', params=params)
            
            posts = []
            for item in data.get('feed', []):
                post = item.get('post', {})
                record = post.get('record', {})
                posts.append({
                    'uri': post.get('uri'),
                    'cid': post.get('cid'),
                    'author': {
                        'did': post.get('author', {}).get('did'),
                        'handle': post.get('author', {}).get('handle'),
                        'display_name': post.get('author', {}).get('displayName'),
                    },
                    'text': record.get('text'),
                    'created_at': record.get('createdAt'),
                    'likes_count': post.get('likeCount', 0),
                    'replies_count': post.get('replyCount', 0),
                })

            return {
                'posts': posts,
                'cursor': data.get('cursor'),
            }
        except Exception as e:
            logger.error(f"Failed to fetch posts: {str(e)}")
            return {'posts': [], 'cursor': None}

    def refresh_token(self, refresh_token):
        """Refresh the access token using the refresh token."""
        try:
            data = self._make_request('POST', '/xrpc/com.atproto.server.refreshSession', 
                                    json={'refreshToken': refresh_token})
            return {
                'access_token': data.get('accessJwt'),
                'refresh_token': data.get('refreshJwt'),
            }
        except Exception as e:
            logger.error(f"Failed to refresh token: {str(e)}")
            return None

def get_federated_client(profile):
    """Get an authenticated AT Protocol client for a profile."""
    if not profile.federation_server or not profile.federation_access_token:
        return None
    return ATProtocolClient(profile.federation_server, profile.federation_access_token)

@login_required
@require_http_methods(["POST"])
def link_federation_account(request):
    """
    Link an existing AT Protocol account to the user's profile.
    Expects: {
        "server": "https://example.com",
        "handle": "user.example.com",
        "did": "did:plc:abcdef123",
        "access_token": "token123",
        "refresh_token": "refresh123"
    }
    """
    try:
        data = json.loads(request.body)
        profile = request.user.profile
        
        # Update federation details
        profile.federation_server = data.get('server')
        profile.federation_handle = data.get('handle')
        profile.federation_did = data.get('did')
        profile.federation_access_token = data.get('access_token')
        profile.federation_refresh_token = data.get('refresh_token')
        profile.save()
        
        return JsonResponse({
            'status': 'success',
            'message': 'Federation account linked successfully'
        })
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)

@login_required
@require_http_methods(["DELETE"])
def unlink_federation_account(request):
    """Remove federation account details from user profile."""
    try:
        profile = request.user.profile
        
        # Clear federation details
        profile.federation_server = None
        profile.federation_handle = None
        profile.federation_did = None
        profile.federation_access_token = None
        profile.federation_refresh_token = None
        profile.save()
        
        return JsonResponse({
            'status': 'success',
            'message': 'Federation account unlinked successfully'
        })
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)

@csrf_exempt
@require_http_methods(["POST"])
def federation_webhook(request):
    """
    Webhook endpoint for receiving updates from federated servers.
    This endpoint should be registered with the federation server.
    """
    try:
        data = json.loads(request.body)
        
        # Verify the webhook signature
        # TODO: Implement signature verification
        
        # Process the webhook data
        event_type = data.get('type')
        if event_type == 'profile.update':
            handle = data.get('handle')
            profile = Profile.objects.get(federation_handle=handle)
            # Update profile with federation data
            # TODO: Implement profile update logic
        
        return JsonResponse({
            'status': 'success',
            'message': 'Webhook processed successfully'
        })
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)

@login_required
@require_http_methods(["GET"])
def get_federated_profile(request, handle):
    """Fetch a profile from a federated server."""
    try:
        profile = request.user.profile
        client = get_federated_client(profile)
        if not client:
            return JsonResponse({
                'status': 'error',
                'message': 'No federation account linked'
            }, status=400)

        federated_profile = client.get_profile(handle)
        if not federated_profile:
            return JsonResponse({
                'status': 'error',
                'message': 'Profile not found'
            }, status=404)

        return JsonResponse({
            'status': 'success',
            'profile': federated_profile
        })
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)

@login_required
@require_http_methods(["GET"])
def get_federated_posts(request, handle=None):
    """Fetch posts from a federated server."""
    try:
        profile = request.user.profile
        client = get_federated_client(profile)
        if not client:
            return JsonResponse({
                'status': 'error',
                'message': 'No federation account linked'
            }, status=400)

        cursor = request.GET.get('cursor')
        limit = int(request.GET.get('limit', 20))
        result = client.get_posts(handle, cursor, limit)

        return JsonResponse({
            'status': 'success',
            'posts': result['posts'],
            'cursor': result['cursor']
        })
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)

@login_required
@require_http_methods(["POST"])
def refresh_federation_token(request):
    """Refresh the federation server access token using the refresh token."""
    try:
        profile = request.user.profile
        if not profile.federation_refresh_token:
            return JsonResponse({
                'status': 'error',
                'message': 'No refresh token available'
            }, status=400)

        client = ATProtocolClient(profile.federation_server)
        result = client.refresh_token(profile.federation_refresh_token)
        
        if result:
            profile.federation_access_token = result['access_token']
            profile.federation_refresh_token = result['refresh_token']
            profile.save()
            
            return JsonResponse({
                'status': 'success',
                'message': 'Token refreshed successfully'
            })
        else:
            return JsonResponse({
                'status': 'error',
                'message': 'Failed to refresh token'
            }, status=400)
            
    except Exception as e:
        return JsonResponse({
            'status': 'error',
            'message': str(e)
        }, status=400)
