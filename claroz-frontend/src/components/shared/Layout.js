import React, { useState } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  AppBar,
  Box,
  Toolbar,
  IconButton,
  Typography,
  Menu,
  Container,
  Avatar,
  Button,
  Tooltip,
  MenuItem,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import { Home as HomeIcon, Feed as FeedIcon } from '@mui/icons-material';
import { useAuth } from '../../context/AuthContext';

function Layout({ children }) {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const [anchorElUser, setAnchorElUser] = useState(null);

  const handleOpenUserMenu = (event) => {
    setAnchorElUser(event.currentTarget);
  };

  const handleCloseUserMenu = () => {
    setAnchorElUser(null);
  };

  const handleMenuClick = (action) => {
    handleCloseUserMenu();
    switch (action) {
      case 'profile':
        navigate(`/profile/${user.id}`);
        break;
      case 'logout':
        logout();
        navigate('/login');
        break;
      default:
        break;
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar 
        position="sticky" 
        elevation={2}
        sx={{
          backgroundColor: 'background.paper',
          color: 'text.primary',
        }}
      >
        <Container maxWidth="lg">
          <Toolbar disableGutters>
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Typography
                variant="h6"
                noWrap
                component={RouterLink}
                to="/"
                sx={{
                  mr: 4,
                  display: 'flex',
                  fontFamily: 'monospace',
                  fontWeight: 700,
                  letterSpacing: '.3rem',
                  color: 'primary.main',
                  textDecoration: 'none',
                  '&:hover': {
                    color: 'primary.dark',
                  },
                }}
              >
                CLAROZ
              </Typography>

              {user && (
                <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 2 }}>
                  <Button
                    component={RouterLink}
                    to="/home"
                    startIcon={<HomeIcon />}
                    color="inherit"
                    sx={{
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    Home
                  </Button>
                  <Button
                    component={RouterLink}
                    to="/"
                    startIcon={<FeedIcon />}
                    color="inherit"
                    sx={{
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    Feed
                  </Button>
                </Box>
              )}
            </Box>

            <Box sx={{ flexGrow: 1 }} />

            {user ? (
              <Box sx={{ flexGrow: 0 }}>
                <Tooltip title="Open settings">
                  <IconButton 
                    onClick={handleOpenUserMenu} 
                    sx={{ 
                      p: 0.5,
                      border: 2,
                      borderColor: 'primary.main',
                      '&:hover': {
                        borderColor: 'primary.dark',
                      },
                    }}
                  >
                    <Avatar 
                      alt={user.username} 
                      src={user.avatar}
                      sx={{ 
                        width: 32,
                        height: 32,
                      }}
                    />
                  </IconButton>
                </Tooltip>
                <Menu
                  sx={{ mt: '45px' }}
                  id="menu-appbar"
                  anchorEl={anchorElUser}
                  anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                  }}
                  keepMounted
                  transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                  }}
                  open={Boolean(anchorElUser)}
                  onClose={handleCloseUserMenu}
                >
                  <MenuItem 
                    onClick={() => handleMenuClick('profile')}
                    sx={{ 
                      minWidth: 150,
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    <Typography textAlign="center">Profile</Typography>
                  </MenuItem>
                  <MenuItem 
                    onClick={() => handleMenuClick('logout')}
                    sx={{
                      color: 'error.main',
                      '&:hover': {
                        backgroundColor: 'error.light',
                      },
                    }}
                  >
                    <Typography textAlign="center">Logout</Typography>
                  </MenuItem>
                </Menu>
              </Box>
            ) : (
              <Box sx={{ flexGrow: 0, display: 'flex', gap: 2 }}>
                <Button
                  component={RouterLink}
                  to="/login"
                  color="primary"
                  variant="text"
                  sx={{
                    fontWeight: 500,
                    '&:hover': {
                      backgroundColor: 'primary.light',
                      color: 'primary.contrastText',
                    },
                  }}
                >
                  Login
                </Button>
                <Button
                  component={RouterLink}
                  to="/register"
                  color="primary"
                  variant="contained"
                  sx={{
                    fontWeight: 500,
                    boxShadow: 2,
                    '&:hover': {
                      boxShadow: 4,
                    },
                  }}
                >
                  Register
                </Button>
              </Box>
            )}
          </Toolbar>
        </Container>
      </AppBar>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          py: { xs: 2, sm: 3 },
          px: { xs: 1, sm: 2, md: 3 },
          bgcolor: 'background.default',
          maxWidth: 'lg',
          mx: 'auto',
          width: '100%',
        }}
      >
        {children}
      </Box>
    </Box>
  );
}

export default Layout;
