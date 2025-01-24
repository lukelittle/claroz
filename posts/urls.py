from django.urls import path
from . import views, federation

app_name = 'posts'

urlpatterns = [
    # Home feed (following)
    path('', views.FeedView.as_view(), name='feed'),
    path('all/', views.PostListView.as_view(), name='list'),
    
    # Post CRUD operations
    path('create/', views.PostCreateView.as_view(), name='create'),
    path('<int:post_id>/', views.PostDetailView.as_view(), name='detail'),
    path('<int:post_id>/edit/', views.PostUpdateView.as_view(), name='edit'),
    path('<int:post_id>/delete/', views.PostDeleteView.as_view(), name='delete'),
    
    # Post interactions
    path('<int:post_id>/like/', views.PostLikeView.as_view(), name='like'),
    path('<int:post_id>/unlike/', views.PostUnlikeView.as_view(), name='unlike'),
    path('<int:post_id>/comment/', views.CommentCreateView.as_view(), name='comment'),
    path('comment/<int:comment_id>/delete/', views.CommentDeleteView.as_view(), name='comment_delete'),
    
    # Profile views
    path('profile/<str:username>/', views.ProfilePostsView.as_view(), name='profile_posts'),
    path('profile/<str:username>/likes/', views.ProfileLikesView.as_view(), name='profile_likes'),
    path('profile/<str:username>/follow/', views.ProfileFollowView.as_view(), name='follow'),
    path('profile/<str:username>/unfollow/', views.ProfileUnfollowView.as_view(), name='unfollow'),
    
    # Federation endpoints
    path('federation/link/', federation.link_federation_account, name='federation_link'),
    path('federation/unlink/', federation.unlink_federation_account, name='federation_unlink'),
    path('federation/webhook/', federation.federation_webhook, name='federation_webhook'),
    path('federation/refresh-token/', federation.refresh_federation_token, name='federation_refresh_token'),
    path('federation/profile/<str:handle>/', federation.get_federated_profile, name='federation_profile'),
    path('federation/posts/', federation.get_federated_posts, name='federation_posts'),
    path('federation/posts/<str:handle>/', federation.get_federated_posts, name='federation_user_posts'),
]
