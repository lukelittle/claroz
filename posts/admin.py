from django.contrib import admin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.contrib.auth.models import User
from .models import Post, Like, Comment, Profile

class ProfileInline(admin.StackedInline):
    model = Profile
    can_delete = False
    verbose_name_plural = 'Profile'

class UserAdmin(BaseUserAdmin):
    inlines = (ProfileInline,)
    list_display = ('username', 'email', 'get_did', 'get_handle', 'get_followers', 'get_following', 'is_staff')
    
    def get_did(self, obj):
        return obj.profile.did if hasattr(obj, 'profile') else ''
    get_did.short_description = 'DID'
    
    def get_handle(self, obj):
        return obj.profile.handle if hasattr(obj, 'profile') else ''
    get_handle.short_description = 'Handle'
    
    def get_followers(self, obj):
        return obj.profile.followers_count if hasattr(obj, 'profile') else 0
    get_followers.short_description = 'Followers'
    
    def get_following(self, obj):
        return obj.profile.follows_count if hasattr(obj, 'profile') else 0
    get_following.short_description = 'Following'

# Re-register UserAdmin
admin.site.unregister(User)
admin.site.register(User, UserAdmin)

@admin.register(Post)
class PostAdmin(admin.ModelAdmin):
    list_display = ('get_username', 'caption', 'created_at', 'likes_count', 'comments_count')
    list_filter = ('created_at', 'profile')
    search_fields = ('profile__user__username', 'caption')
    date_hierarchy = 'created_at'

    def get_username(self, obj):
        return obj.profile.user.username
    get_username.short_description = 'Username'
    get_username.admin_order_field = 'profile__user__username'

    def likes_count(self, obj):
        return obj.likes.count()
    likes_count.short_description = 'Likes'

    def comments_count(self, obj):
        return obj.comments.count()
    comments_count.short_description = 'Comments'

@admin.register(Like)
class LikeAdmin(admin.ModelAdmin):
    list_display = ('get_username', 'post', 'created_at')
    list_filter = ('created_at', 'profile')
    search_fields = ('profile__user__username', 'post__caption')

    def get_username(self, obj):
        return obj.profile.user.username
    get_username.short_description = 'Username'
    get_username.admin_order_field = 'profile__user__username'

@admin.register(Comment)
class CommentAdmin(admin.ModelAdmin):
    list_display = ('get_username', 'post', 'text', 'created_at')
    list_filter = ('created_at', 'profile')
    search_fields = ('profile__user__username', 'text', 'post__caption')

    def get_username(self, obj):
        return obj.profile.user.username
    get_username.short_description = 'Username'
    get_username.admin_order_field = 'profile__user__username'
