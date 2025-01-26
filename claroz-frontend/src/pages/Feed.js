import React, { useState, useEffect, useCallback } from 'react';
import {
  Container,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Button,
  Divider,
} from '@mui/material';
import { Refresh as RefreshIcon } from '@mui/icons-material';
import Layout from '../components/shared/Layout';
import PostCard from '../components/posts/PostCard';
import FederatedContent from '../components/federation/FederatedContent';
import { postAPI, federationAPI } from '../services/api';

function Feed() {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshing, setRefreshing] = useState(false);

  const fetchAllPosts = useCallback(async () => {
    try {
      setLoading(true);
      // Fetch both local and federated posts
      const [localResponse, federatedResponse] = await Promise.all([
        postAPI.getPosts(),
        federationAPI.getFederatedPosts(),
      ]);

      // Merge and sort posts by creation date
      const allPosts = [
        ...localResponse.data,
        ...federatedResponse.data,
      ].sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

      setPosts(allPosts);
    } catch (err) {
      setError('Failed to fetch posts');
      console.error('Error fetching posts:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAllPosts();
  }, [fetchAllPosts]);

  const handleRefresh = async () => {
    try {
      setRefreshing(true);
      await fetchAllPosts();
    } finally {
      setRefreshing(false);
    }
  };

  const handlePostUpdate = async (postId, isFederated = false) => {
    try {
      let updatedPost;
      if (isFederated) {
        const response = await federationAPI.getFederatedPosts(postId);
        updatedPost = response.data[0];
      } else {
        const response = await postAPI.getPost(postId);
        updatedPost = response.data;
      }

      setPosts((prevPosts) =>
        prevPosts.map((post) =>
          post.id === postId ? updatedPost : post
        )
      );
    } catch (err) {
      console.error('Error updating post:', err);
    }
  };

  if (loading && !refreshing) {
    return (
      <Layout>
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '50vh',
          }}
        >
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  const handleProfileResolved = async (profileData) => {
    try {
      // Fetch posts for the newly resolved profile
      const response = await federationAPI.getFederatedPosts(profileData.did);
      // Add new posts to the feed
      setPosts((prevPosts) => {
        const newPosts = [...response.data, ...prevPosts];
        return newPosts.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
      });
    } catch (err) {
      setError('Failed to fetch federated posts');
      console.error('Error fetching federated posts:', err);
    }
  };

  return (
    <Layout>
      <Container maxWidth="md">
        <FederatedContent onProfileResolved={handleProfileResolved} />
        <Divider sx={{ mb: 3 }} />
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            mb: 3,
          }}
        >
          <Typography variant="h5" component="h1">
            Feed
          </Typography>
          <Button
            startIcon={refreshing ? <CircularProgress size={20} /> : <RefreshIcon />}
            onClick={handleRefresh}
            disabled={refreshing}
          >
            Refresh
          </Button>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {posts.length === 0 ? (
          <Typography variant="body1" color="text.secondary" align="center">
            No posts yet. Follow some users to see their posts here!
          </Typography>
        ) : (
          posts.map((post) => (
            <PostCard
              key={post.id}
              post={post}
              onPostUpdate={handlePostUpdate}
            />
          ))
        )}
      </Container>
    </Layout>
  );
}

export default Feed;
