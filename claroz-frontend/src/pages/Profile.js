import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import {
  Container,
  Grid,
  Card,
  CardContent,
  Typography,
  Avatar,
  Button,
  Box,
  CircularProgress,
  Alert,
  Tabs,
  Tab,
  Chip,
  Divider,
} from '@mui/material';
import {
  Edit as EditIcon,
  Sync as SyncIcon,
  PersonAdd as PersonAddIcon,
  PersonRemove as PersonRemoveIcon,
} from '@mui/icons-material';
import Layout from '../components/shared/Layout';
import PostCard from '../components/posts/PostCard';
import EditProfileDialog from '../components/profile/EditProfileDialog';
import FollowList from '../components/profile/FollowList';
import { userAPI, postAPI, federationAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

function Profile() {
  const { userId } = useParams();
  const { user: currentUser } = useAuth();
  const [user, setUser] = useState(null);
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isFollowing, setIsFollowing] = useState(false);
  const [syncingProfile, setSyncingProfile] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [followListOpen, setFollowListOpen] = useState(false);
  const [followListType, setFollowListType] = useState('followers');
  const [activeTab, setActiveTab] = useState(0);

  const fetchUserData = useCallback(async () => {
    try {
      setLoading(true);
      let userResponse, postsResponse;

      // Check if this is a federated profile
      if (userId.startsWith('did:')) {
        userResponse = await federationAPI.getFederatedProfile(userId);
        postsResponse = await federationAPI.getFederatedPosts(userId);
      } else {
        [userResponse, postsResponse] = await Promise.all([
          userAPI.getProfile(userId),
          postAPI.getPosts(`?userId=${userId}`),
        ]);
      }

      setUser(userResponse.data);
      setPosts(postsResponse.data);
      
      // Only check following status for local users
      if (!userId.startsWith('did:')) {
        setIsFollowing(userResponse.data.followers?.some(
          (follower) => follower.id === currentUser.id
        ));
      }
    } catch (err) {
      setError('Failed to fetch user data');
      console.error('Error fetching user data:', err);
    } finally {
      setLoading(false);
    }
  }, [userId, currentUser.id]);

  useEffect(() => {
    fetchUserData();
  }, [fetchUserData]);

  const handleFollowToggle = async () => {
    try {
      if (isFollowing) {
        await userAPI.unfollowUser(userId);
      } else {
        await userAPI.followUser(userId);
      }
      setIsFollowing(!isFollowing);
      // Refresh user data to update followers count
      fetchUserData();
    } catch (err) {
      setError('Failed to update follow status');
      console.error('Error updating follow status:', err);
    }
  };

  const handleSyncProfile = async () => {
    if (!user?.did) return;

    try {
      setSyncingProfile(true);
      await federationAPI.syncProfile(user.did);
      
      // Fetch updated federated data
      const [updatedProfile, updatedPosts] = await Promise.all([
        federationAPI.getFederatedProfile(user.did),
        federationAPI.getFederatedPosts(user.did),
      ]);
      
      setUser(updatedProfile.data);
      setPosts(updatedPosts.data);
    } catch (err) {
      setError('Failed to sync federated profile');
      console.error('Error syncing profile:', err);
    } finally {
      setSyncingProfile(false);
    }
  };

  const handleEditProfile = async (formData) => {
    try {
      await userAPI.updateProfile(formData);
      await fetchUserData();
    } catch (err) {
      throw new Error('Failed to update profile');
    }
  };

  const handleShowFollowList = (type) => {
    setFollowListType(type);
    setFollowListOpen(true);
  };

  const handlePostUpdate = async (postId) => {
    try {
      let response;
      if (user.federationType === 'remote') {
        response = await federationAPI.getFederatedPosts(user.did);
        const updatedPost = response.data.find(p => p.id === postId);
        if (updatedPost) {
          setPosts((prevPosts) =>
            prevPosts.map((post) =>
              post.id === postId ? updatedPost : post
            )
          );
        }
      } else {
        response = await postAPI.getPost(postId);
        setPosts((prevPosts) =>
          prevPosts.map((post) =>
            post.id === postId ? response.data : post
          )
        );
      }
    } catch (err) {
      console.error('Error updating post:', err);
    }
  };

  if (loading) {
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

  if (!user) {
    return (
      <Layout>
        <Container>
          <Alert severity="error">User not found</Alert>
        </Container>
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="md">
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Card 
          elevation={2}
          sx={{ 
            mb: 3,
            borderRadius: 2,
            background: 'linear-gradient(to bottom, primary.light, background.paper)',
            backgroundSize: '100% 150px',
            backgroundRepeat: 'no-repeat'
          }}
        >
          <CardContent sx={{ pt: 3 }}>
            <Box
              sx={{
                display: 'flex',
                alignItems: 'flex-start',
                mb: 2,
              }}
            >
              <Avatar
                src={user.avatar}
                alt={user.username}
                sx={{ 
                  width: { xs: 100, sm: 120, md: 140 },
                  height: { xs: 100, sm: 120, md: 140 },
                  mr: 3,
                  border: 4,
                  borderColor: 'background.paper',
                  boxShadow: 3
                }}
              />
              <Box sx={{ flexGrow: 1 }}>
                <Box
                  sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    mb: 1,
                  }}
                >
                  <Box>
                    <Typography variant="h4" sx={{ 
                      fontWeight: 600,
                      color: 'text.primary',
                      mb: 0.5
                    }}>
                      {user.fullName}
                    </Typography>
                    <Typography 
                      variant="subtitle1" 
                      sx={{ 
                        color: 'text.secondary',
                        display: 'flex',
                        alignItems: 'center',
                        flexWrap: 'wrap',
                        gap: 1
                      }}
                    >
                      @{user.username}
                      {user.federationType === 'remote' && (
                        <Chip
                          label={user.handle}
                          size="small"
                          color="primary"
                          variant="outlined"
                          sx={{ 
                            borderRadius: 4,
                            backgroundColor: 'background.paper',
                            '& .MuiChip-label': {
                              px: 1
                            }
                          }}
                        />
                      )}
                    </Typography>
                  </Box>
                  <Box>
                    {user.federationType === 'remote' ? (
                      <Button
                        variant="contained"
                        onClick={handleSyncProfile}
                        disabled={syncingProfile}
                        startIcon={syncingProfile ? <CircularProgress size={20} /> : <SyncIcon />}
                        sx={{
                          borderRadius: 2,
                          textTransform: 'none',
                          boxShadow: 2,
                          '&:hover': {
                            boxShadow: 4
                          }
                        }}
                      >
                        Sync Profile
                      </Button>
                    ) : currentUser.id === userId ? (
                      <Button
                        variant="contained"
                        startIcon={<EditIcon />}
                        onClick={() => setEditDialogOpen(true)}
                        sx={{
                          borderRadius: 2,
                          textTransform: 'none',
                          boxShadow: 2,
                          '&:hover': {
                            boxShadow: 4
                          }
                        }}
                      >
                        Edit Profile
                      </Button>
                    ) : (
                      <Button
                        variant={isFollowing ? 'outlined' : 'contained'}
                        onClick={handleFollowToggle}
                        startIcon={
                          isFollowing ? <PersonRemoveIcon /> : <PersonAddIcon />
                        }
                      >
                        {isFollowing ? 'Unfollow' : 'Follow'}
                      </Button>
                    )}
                    {user.federationType === 'remote' && (
                      <Button
                        variant="outlined"
                        onClick={handleSyncProfile}
                        disabled={syncingProfile}
                        startIcon={
                          syncingProfile ? (
                            <CircularProgress size={20} />
                          ) : (
                            <SyncIcon />
                          )
                        }
                        sx={{ ml: 1 }}
                      >
                        Sync
                      </Button>
                    )}
                  </Box>
                </Box>

                {user.bio && (
                  <Typography variant="body1" sx={{ mt: 2, mb: 2 }}>
                    {user.bio}
                  </Typography>
                )}

                <Box sx={{ 
                  display: 'flex', 
                  gap: 4,
                  mt: 3,
                  borderTop: 1,
                  borderColor: 'divider',
                  pt: 2
                }}>
                  <Button
                    variant="text"
                    onClick={() => handleShowFollowList('followers')}
                  >
                    <Box sx={{ textAlign: 'center' }}>
                      <Typography 
                        variant="h6" 
                        component="span"
                        sx={{ 
                          fontWeight: 'bold',
                          color: 'primary.main'
                        }}
                      >
                        {user.followers?.length || 0}
                      </Typography>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ 
                          display: 'block',
                          mt: 0.5
                        }}
                      >
                        Followers
                      </Typography>
                    </Box>
                  </Button>
                  <Button
                    variant="text"
                    onClick={() => handleShowFollowList('following')}
                  >
                    <Box sx={{ textAlign: 'center' }}>
                      <Typography 
                        variant="h6" 
                        component="span"
                        sx={{ 
                          fontWeight: 'bold',
                          color: 'primary.main'
                        }}
                      >
                        {user.following?.length || 0}
                      </Typography>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ 
                          display: 'block',
                          mt: 0.5
                        }}
                      >
                        Following
                      </Typography>
                    </Box>
                  </Button>
                </Box>
              </Box>
            </Box>

            <Divider sx={{ my: 2 }} />

            <Tabs
              value={activeTab}
              onChange={(e, newValue) => setActiveTab(newValue)}
              centered
              sx={{
                '& .MuiTab-root': {
                  textTransform: 'none',
                  fontWeight: 500,
                  fontSize: '1rem',
                  minWidth: 120
                },
                '& .Mui-selected': {
                  color: 'primary.main'
                },
                '& .MuiTabs-indicator': {
                  height: 3,
                  borderRadius: 1.5
                }
              }}
            >
              <Tab label="Posts" />
              <Tab label="Likes" />
            </Tabs>
          </CardContent>
        </Card>

        <Grid container spacing={3}>
          {activeTab === 0 &&
            posts.map((post) => (
              <Grid item xs={12} key={post.id}>
                <PostCard post={post} onPostUpdate={handlePostUpdate} />
              </Grid>
            ))}
          {posts.length === 0 && (
            <Grid item xs={12}>
              <Typography variant="body2" color="text.secondary" align="center">
                No posts yet
              </Typography>
            </Grid>
          )}
        </Grid>

        <EditProfileDialog
          open={editDialogOpen}
          onClose={() => setEditDialogOpen(false)}
          user={user}
          onSave={handleEditProfile}
        />

        <FollowList
          open={followListOpen}
          onClose={() => setFollowListOpen(false)}
          title={followListType === 'followers' ? 'Followers' : 'Following'}
          users={followListType === 'followers' ? user.followers : user.following}
        />
      </Container>
    </Layout>
  );
}

export default Profile;
