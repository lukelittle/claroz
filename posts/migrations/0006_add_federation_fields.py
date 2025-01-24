# Generated by Django 5.1.5 on 2025-01-24 17:31

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('posts', '0005_create_missing_profiles'),
    ]

    operations = [
        migrations.AddField(
            model_name='profile',
            name='federation_access_token',
            field=models.CharField(blank=True, help_text='Access token for the federated server', max_length=255, null=True),
        ),
        migrations.AddField(
            model_name='profile',
            name='federation_did',
            field=models.CharField(blank=True, help_text='DID on the federated server', max_length=255, null=True),
        ),
        migrations.AddField(
            model_name='profile',
            name='federation_handle',
            field=models.CharField(blank=True, help_text='Handle on the federated server', max_length=255, null=True),
        ),
        migrations.AddField(
            model_name='profile',
            name='federation_refresh_token',
            field=models.CharField(blank=True, help_text='Refresh token for the federated server', max_length=255, null=True),
        ),
        migrations.AddField(
            model_name='profile',
            name='federation_server',
            field=models.URLField(blank=True, help_text='URL of the federated identity server', max_length=255, null=True),
        ),
    ]
