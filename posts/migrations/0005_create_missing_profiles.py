from django.db import migrations

def create_missing_profiles(apps, schema_editor):
    User = apps.get_model('auth', 'User')
    Profile = apps.get_model('posts', 'Profile')
    
    for user in User.objects.all():
        Profile.objects.get_or_create(
            user=user,
            defaults={
                'did': None,
                'handle': None,
                'display_name': user.username,
            }
        )

def reverse_profiles(apps, schema_editor):
    pass

class Migration(migrations.Migration):

    dependencies = [
        ('posts', '0004_populate_profile_fields'),
    ]

    operations = [
        migrations.RunPython(
            create_missing_profiles,
            reverse_profiles
        ),
    ]
