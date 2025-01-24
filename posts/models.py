from django.db import models
from django.contrib.auth.models import User
from django.db.models.signals import post_save
from django.dispatch import receiver

class Profile(models.Model):
    user = models.OneToOneField(User, on_delete=models.CASCADE)
    followers = models.ManyToManyField('self', symmetrical=False, related_name='following', blank=True)
    did = models.CharField(max_length=255, unique=True, blank=True, null=True)
    profile_picture = models.ImageField(upload_to='profiles/', storage='posts.ipfs.IPFSStorage', blank=True, null=True)
    profile_picture_cid = models.CharField(max_length=64, blank=True, null=True, help_text="IPFS Content Identifier for the profile picture")
    profile_picture_filename = models.CharField(max_length=255, blank=True, null=True, help_text="Original filename of the profile picture")
    bio = models.TextField(max_length=500, blank=True)
    website = models.URLField(max_length=200, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    
    # AT Protocol specific fields
    handle = models.CharField(max_length=255, unique=True, blank=True, null=True)
    display_name = models.CharField(max_length=64, blank=True)
    follows_count = models.PositiveIntegerField(default=0)
    followers_count = models.PositiveIntegerField(default=0)
    posts_count = models.PositiveIntegerField(default=0)
    
    # Federation fields
    federation_server = models.URLField(max_length=255, blank=True, null=True, help_text="URL of the federated identity server")
    federation_handle = models.CharField(max_length=255, blank=True, null=True, help_text="Handle on the federated server")
    federation_did = models.CharField(max_length=255, blank=True, null=True, help_text="DID on the federated server")
    federation_access_token = models.CharField(max_length=255, blank=True, null=True, help_text="Access token for the federated server")
    federation_refresh_token = models.CharField(max_length=255, blank=True, null=True, help_text="Refresh token for the federated server")
    
    def get_ipfs_url(self):
        """Get the IPFS gateway URL for the profile picture."""
        from django.conf import settings
        if self.profile_picture_cid:
            gateway_url = getattr(settings, 'IPFS_GATEWAY_URL', 'http://localhost:8080/ipfs')
            return f"{gateway_url}/{self.profile_picture_cid}"
        return None

    def save(self, *args, **kwargs):
        # Handle IPFS storage for profile picture
        if self.profile_picture and not self.profile_picture_cid:
            # Store original filename
            self.profile_picture_filename = self.profile_picture.name
            
            # Use storage backend to get CID
            storage = self.profile_picture.storage
            filename = storage._save(self.profile_picture.name, self.profile_picture)
            self.profile_picture_cid = filename  # Storage returns CID as filename
            
            # Clear profile picture field since we'll use CID
            self.profile_picture = None
        
        super().save(*args, **kwargs)

    def __str__(self):
        return f"{self.user.username}'s profile"
    
    def follow(self, profile_to_follow):
        """Follow another profile if not already following"""
        if profile_to_follow != self:
            self.following.add(profile_to_follow)
    
    def unfollow(self, profile_to_unfollow):
        """Unfollow another profile if currently following"""
        self.following.remove(profile_to_unfollow)
    
    def is_following(self, profile):
        """Check if this profile is following another profile"""
        return self.following.filter(pk=profile.pk).exists()

@receiver(post_save, sender=User)
def create_user_profile(sender, instance, created, **kwargs):
    if created:
        Profile.objects.create(user=instance)

@receiver(post_save, sender=User)
def save_user_profile(sender, instance, **kwargs):
    instance.profile.save()

@receiver(models.signals.m2m_changed, sender=Profile.followers.through)
def update_follower_counts(sender, instance, action, reverse, model, pk_set, **kwargs):
    """Update follower and following counts when relationships change"""
    if action in ["post_add", "post_remove"]:
        for pk in pk_set:
            # Update the followed profile's followers count
            followed_profile = Profile.objects.get(pk=pk)
            followed_profile.followers_count = followed_profile.followers.count()
            followed_profile.save()
            
            # Update the following profile's follows count
            following_profile = instance if not reverse else Profile.objects.get(pk=pk)
            following_profile.follows_count = following_profile.following.count()
            following_profile.save()

class Post(models.Model):
    profile = models.ForeignKey(Profile, on_delete=models.CASCADE, related_name='posts', null=True)
    user = models.ForeignKey(User, on_delete=models.CASCADE)  # Temporary field for migration
    image = models.ImageField(upload_to='posts/', storage='posts.ipfs.IPFSStorage', null=True, blank=True)
    image_cid = models.CharField(max_length=64, help_text="IPFS Content Identifier for the image")
    original_filename = models.CharField(max_length=255, blank=True, help_text="Original filename of the uploaded image")
    caption = models.TextField(blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ['-created_at']

    def __str__(self):
        username = self.profile.user.username if self.profile else self.user.username
        return f'{username}\'s post - {self.created_at}'

    def get_ipfs_url(self):
        """Get the IPFS gateway URL for the image."""
        from django.conf import settings
        if self.image_cid:
            gateway_url = getattr(settings, 'IPFS_GATEWAY_URL', 'http://localhost:8080/ipfs')
            return f"{gateway_url}/{self.image_cid}"
        return None

    def save(self, *args, **kwargs):
        if not self.profile and self.user:
            self.profile = self.user.profile
        
        # Handle IPFS storage for new images
        if self.image and not self.image_cid:
            # Store original filename
            self.original_filename = self.image.name
            
            # Use storage backend to get CID
            storage = self.image.storage
            filename = storage._save(self.image.name, self.image)
            self.image_cid = filename  # Storage returns CID as filename
            
            # Clear image field since we'll use CID
            self.image = None
        
        super().save(*args, **kwargs)
        
        # Update post count
        if self.profile:
            self.profile.posts_count = self.profile.posts.count()
            self.profile.save()

@receiver(models.signals.post_delete, sender=Post)
def update_profile_posts_count(sender, instance, **kwargs):
    instance.profile.posts_count = instance.profile.posts.count()
    instance.profile.save()

class Like(models.Model):
    profile = models.ForeignKey(Profile, on_delete=models.CASCADE, null=True)
    user = models.ForeignKey(User, on_delete=models.CASCADE)  # Temporary field for migration
    post = models.ForeignKey(Post, on_delete=models.CASCADE, related_name='likes')
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        unique_together = ('user', 'post')  # We'll update this in a later migration

    def __str__(self):
        if self.profile:
            return f'{self.profile.user.username} likes {self.post}'
        return f'{self.user.username} likes {self.post}'

    def save(self, *args, **kwargs):
        if not self.profile and self.user:
            self.profile = self.user.profile
        super().save(*args, **kwargs)

class Comment(models.Model):
    profile = models.ForeignKey(Profile, on_delete=models.CASCADE, null=True)
    user = models.ForeignKey(User, on_delete=models.CASCADE)  # Temporary field for migration
    post = models.ForeignKey(Post, on_delete=models.CASCADE, related_name='comments')
    text = models.TextField()
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ['created_at']

    def __str__(self):
        if self.profile:
            return f'{self.profile.user.username} commented on {self.post}'
        return f'{self.user.username} commented on {self.post}'

    def save(self, *args, **kwargs):
        if not self.profile and self.user:
            self.profile = self.user.profile
        super().save(*args, **kwargs)
