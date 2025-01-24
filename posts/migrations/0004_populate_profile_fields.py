from django.db import migrations

def populate_profile_fields(apps, schema_editor):
    Post = apps.get_model('posts', 'Post')
    Like = apps.get_model('posts', 'Like')
    Comment = apps.get_model('posts', 'Comment')

    # Update Post profiles
    for post in Post.objects.all():
        if post.user and not post.profile:
            post.profile = post.user.profile
            post.save()

    # Update Like profiles
    for like in Like.objects.all():
        if like.user and not like.profile:
            like.profile = like.user.profile
            like.save()

    # Update Comment profiles
    for comment in Comment.objects.all():
        if comment.user and not comment.profile:
            comment.profile = comment.user.profile
            comment.save()

def reverse_populate_profile_fields(apps, schema_editor):
    # No need to reverse this migration since we're keeping user fields for now
    pass

class Migration(migrations.Migration):

    dependencies = [
        ('posts', '0003_comment_profile_like_profile_post_profile_and_more'),
    ]

    operations = [
        migrations.RunPython(
            populate_profile_fields,
            reverse_populate_profile_fields
        ),
    ]
