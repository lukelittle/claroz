# Generated by Django 5.1.5 on 2025-01-24 18:46

from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('posts', '0009_add_profile_ipfs_fields'),
    ]

    operations = [
        migrations.AlterField(
            model_name='post',
            name='image',
            field=models.ImageField(blank=True, null=True, storage='posts.ipfs.IPFSStorage', upload_to='posts/'),
        ),
        migrations.AlterField(
            model_name='post',
            name='image_cid',
            field=models.CharField(help_text='IPFS Content Identifier for the image', max_length=64),
        ),
    ]
