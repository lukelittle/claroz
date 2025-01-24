from django.views.generic import ListView, DetailView, CreateView, UpdateView, DeleteView
from django.db.models import Q
from django.views.generic.edit import FormMixin
from django.contrib.auth.mixins import LoginRequiredMixin, UserPassesTestMixin
from django.views.generic.base import View
from django.shortcuts import get_object_or_404, redirect
from django.urls.base import reverse_lazy
import uuid
from .forms import SignUpWithFederationForm
from django.http import JsonResponse, Http404
from django.views.decorators.http import require_POST
from django.utils.decorators import method_decorator
from django.contrib.auth.models import User
from django.urls import reverse
from django.contrib import messages
from django.db.models import Prefetch
from .models import Post, Like, Comment

class SignUpView(CreateView):
    form_class = SignUpWithFederationForm
    template_name = 'registration/signup.html'
    success_url = reverse_lazy('login')

    def form_valid(self, form):
        response = super().form_valid(form)
        user = self.object
        
        # Generate local DID and set handle if no federation DID provided
        profile = user.profile
        if not profile.federation_did:
            profile.did = f"did:web:{uuid.uuid4()}"
            profile.handle = f"@{user.username}"
            profile.display_name = user.username
            profile.save()
        
        messages.success(self.request, 'Account created successfully! Please log in.')
        return response

class PostListView(ListView):
    model = Post
    template_name = 'posts/post_list.html'
    context_object_name = 'posts'
    paginate_by = 10
    ordering = ['-created_at']

    def get_queryset(self):
        return Post.objects.select_related('profile__user').prefetch_related('likes', 'comments')

class PostDetailView(DetailView):
    model = Post
    template_name = 'posts/post_detail.html'
    context_object_name = 'post'
    pk_url_kwarg = 'post_id'

    def get_queryset(self):
        return Post.objects.select_related('profile__user').prefetch_related('likes', 'comments')

class PostCreateView(LoginRequiredMixin, CreateView):
    model = Post
    template_name = 'posts/post_create.html'
    fields = ['image', 'caption']

    def form_valid(self, form):
        form.instance.profile = self.request.user.profile
        response = super().form_valid(form)
        messages.success(self.request, 'Post created successfully!')
        return response

    def get_success_url(self):
        return reverse('posts:detail', kwargs={'post_id': self.object.id})

class PostUpdateView(LoginRequiredMixin, UserPassesTestMixin, UpdateView):
    model = Post
    template_name = 'posts/post_edit.html'
    fields = ['caption']
    pk_url_kwarg = 'post_id'

    def test_func(self):
        post = self.get_object()
        return post.profile == self.request.user.profile

    def form_valid(self, form):
        response = super().form_valid(form)
        messages.success(self.request, 'Post updated successfully!')
        return response

    def get_success_url(self):
        return reverse('posts:detail', kwargs={'post_id': self.object.id})

class PostDeleteView(LoginRequiredMixin, UserPassesTestMixin, DeleteView):
    model = Post
    template_name = 'posts/post_delete.html'
    success_url = reverse_lazy('posts:list')
    pk_url_kwarg = 'post_id'

    def test_func(self):
        post = self.get_object()
        return post.profile == self.request.user.profile

    def delete(self, request, *args, **kwargs):
        messages.success(self.request, 'Post deleted successfully!')
        return super().delete(request, *args, **kwargs)

class PostLikeView(LoginRequiredMixin, View):
    def post(self, request, post_id):
        post = get_object_or_404(Post, id=post_id)
        like, created = Like.objects.get_or_create(
            profile=request.user.profile,
            post=post
        )
        return JsonResponse({
            'liked': True,
            'likes_count': post.likes.count()
        })

class PostUnlikeView(LoginRequiredMixin, View):
    def post(self, request, post_id):
        post = get_object_or_404(Post, id=post_id)
        Like.objects.filter(
            profile=request.user.profile,
            post=post
        ).delete()
        return JsonResponse({
            'liked': False,
            'likes_count': post.likes.count()
        })

class CommentCreateView(LoginRequiredMixin, CreateView):
    model = Comment
    fields = ['text']
    http_method_names = ['post']

    def form_valid(self, form):
        post = get_object_or_404(Post, id=self.kwargs['post_id'])
        form.instance.profile = self.request.user.profile
        form.instance.post = post
        response = super().form_valid(form)
        messages.success(self.request, 'Comment added successfully!')
        return response

    def form_invalid(self, form):
        messages.error(self.request, 'Comment text is required!')
        return redirect('posts:detail', post_id=self.kwargs['post_id'])

    def get_success_url(self):
        return reverse('posts:detail', kwargs={'post_id': self.kwargs['post_id']})

class CommentDeleteView(LoginRequiredMixin, UserPassesTestMixin, DeleteView):
    model = Comment
    pk_url_kwarg = 'comment_id'

    def test_func(self):
        comment = self.get_object()
        return comment.profile == self.request.user.profile

    def get_success_url(self):
        return reverse('posts:detail', kwargs={'post_id': self.object.post.id})

    def delete(self, request, *args, **kwargs):
        messages.success(self.request, 'Comment deleted successfully!')
        return super().delete(request, *args, **kwargs)

class ProfilePostsView(ListView):
    model = Post
    template_name = 'posts/profile_posts.html'
    context_object_name = 'posts'
    paginate_by = 10

    def get_queryset(self):
        self.profile_user = get_object_or_404(User, username=self.kwargs['username'])
        profile = self.profile_user.profile
        
        # Get local posts
        local_posts = Post.objects.filter(
            profile=profile
        ).select_related('profile__user').prefetch_related('likes', 'comments')

        # Get federated posts if the profile has federation enabled
        federated_posts = []
        if profile.federation_server and profile.federation_access_token:
            from .federation import get_federated_client
            client = get_federated_client(profile)
            if client and profile.federation_handle:
                result = client.get_posts(profile.federation_handle, limit=10)
                if result and result['posts']:
                    federated_posts = result['posts']

        # Convert federated posts to a format compatible with the template
        from django.utils.timezone import datetime, make_aware
        from django.utils.safestring import mark_safe
        import pytz

        formatted_federated_posts = []
        for post in federated_posts:
            created_at = make_aware(datetime.fromisoformat(post['created_at'].replace('Z', '+00:00')))
            formatted_federated_posts.append({
                'id': post['uri'],
                'is_federated': True,
                'profile': {
                    'user': {
                        'username': post['author']['handle'],
                    },
                    'federation_handle': post['author']['handle'],
                    'display_name': post['author']['display_name'],
                },
                'image': None,  # Federated posts might not have images
                'caption': mark_safe(post['text']),
                'created_at': created_at,
                'likes_count': post['likes_count'],
                'comments_count': post['replies_count'],
            })

        # Combine and sort all posts by creation time
        from itertools import chain
        all_posts = list(chain(local_posts, formatted_federated_posts))
        return sorted(all_posts, key=lambda x: x['created_at'] if isinstance(x, dict) else x.created_at, reverse=True)

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context['profile_user'] = self.profile_user
        context['federation_enabled'] = bool(
            self.profile_user.profile.federation_server and 
            self.profile_user.profile.federation_access_token
        )
        if self.request.user.is_authenticated and self.request.user != self.profile_user:
            context['is_following'] = self.request.user.profile.is_following(self.profile_user.profile)
        return context

class ProfileFollowView(LoginRequiredMixin, View):
    def post(self, request, username):
        user_to_follow = get_object_or_404(User, username=username)
        if request.user == user_to_follow:
            return JsonResponse({'error': 'You cannot follow yourself'}, status=400)
        
        request.user.profile.follow(user_to_follow.profile)
        return JsonResponse({
            'following': True,
            'followers_count': user_to_follow.profile.followers_count,
            'follows_count': request.user.profile.follows_count
        })

class ProfileUnfollowView(LoginRequiredMixin, View):
    def post(self, request, username):
        user_to_unfollow = get_object_or_404(User, username=username)
        if request.user == user_to_unfollow:
            return JsonResponse({'error': 'You cannot unfollow yourself'}, status=400)
        
        request.user.profile.unfollow(user_to_unfollow.profile)
        return JsonResponse({
            'following': False,
            'followers_count': user_to_unfollow.profile.followers_count,
            'follows_count': request.user.profile.follows_count
        })

class FeedView(LoginRequiredMixin, ListView):
    model = Post
    template_name = 'posts/feed.html'
    context_object_name = 'posts'
    paginate_by = 10

    def get_queryset(self):
        # Get local posts from followed profiles
        following_profiles = self.request.user.profile.following.all()
        local_posts = Post.objects.filter(
            profile__in=following_profiles
        ).select_related('profile__user').prefetch_related('likes', 'comments')

        # Get federated posts if user has federation enabled
        federated_posts = []
        if self.request.user.profile.federation_server and self.request.user.profile.federation_access_token:
            from .federation import get_federated_client
            client = get_federated_client(self.request.user.profile)
            if client:
                # Get posts from followed federated profiles
                for profile in following_profiles:
                    if profile.federation_handle:
                        result = client.get_posts(profile.federation_handle, limit=10)
                        if result and result['posts']:
                            federated_posts.extend(result['posts'])

        # Convert federated posts to a format compatible with the template
        from django.utils.timezone import datetime, make_aware
        from django.utils.safestring import mark_safe
        import pytz

        formatted_federated_posts = []
        for post in federated_posts:
            created_at = make_aware(datetime.fromisoformat(post['created_at'].replace('Z', '+00:00')))
            formatted_federated_posts.append({
                'id': post['uri'],
                'is_federated': True,
                'profile': {
                    'user': {
                        'username': post['author']['handle'],
                    },
                    'federation_handle': post['author']['handle'],
                    'display_name': post['author']['display_name'],
                },
                'image': None,  # Federated posts might not have images
                'caption': mark_safe(post['text']),
                'created_at': created_at,
                'likes_count': post['likes_count'],
                'comments_count': post['replies_count'],
            })

        # Combine and sort all posts by creation time
        from itertools import chain
        all_posts = list(chain(local_posts, formatted_federated_posts))
        return sorted(all_posts, key=lambda x: x['created_at'] if isinstance(x, dict) else x.created_at, reverse=True)

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context['federation_enabled'] = bool(
            self.request.user.profile.federation_server and 
            self.request.user.profile.federation_access_token
        )
        return context

class ProfileLikesView(ListView):
    model = Post
    template_name = 'posts/profile_likes.html'
    context_object_name = 'posts'
    paginate_by = 10

    def get_queryset(self):
        self.profile_user = get_object_or_404(User, username=self.kwargs['username'])
        profile = self.profile_user.profile

        # Get local liked posts
        local_posts = Post.objects.filter(
            likes__profile=profile
        ).select_related('profile__user').prefetch_related('likes', 'comments')

        # Get federated liked posts if the profile has federation enabled
        federated_posts = []
        if profile.federation_server and profile.federation_access_token:
            from .federation import get_federated_client
            client = get_federated_client(profile)
            if client and profile.federation_handle:
                # TODO: Implement federation API endpoint for liked posts
                # For now, we'll only show local likes
                pass

        # Convert federated posts to a format compatible with the template
        formatted_federated_posts = []
        for post in federated_posts:
            created_at = make_aware(datetime.fromisoformat(post['created_at'].replace('Z', '+00:00')))
            formatted_federated_posts.append({
                'id': post['uri'],
                'is_federated': True,
                'profile': {
                    'user': {
                        'username': post['author']['handle'],
                    },
                    'federation_handle': post['author']['handle'],
                    'display_name': post['author']['display_name'],
                },
                'image': None,
                'caption': mark_safe(post['text']),
                'created_at': created_at,
                'likes_count': post['likes_count'],
                'comments_count': post['replies_count'],
            })

        # Combine and sort all posts by creation time
        from itertools import chain
        all_posts = list(chain(local_posts, formatted_federated_posts))
        return sorted(all_posts, key=lambda x: x['created_at'] if isinstance(x, dict) else x.created_at, reverse=True)

    def get_context_data(self, **kwargs):
        context = super().get_context_data(**kwargs)
        context['profile_user'] = self.profile_user
        context['federation_enabled'] = bool(
            self.profile_user.profile.federation_server and 
            self.profile_user.profile.federation_access_token
        )
        return context
