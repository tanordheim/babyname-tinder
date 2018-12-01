# Babyname Tinder

What do you do when you're about to have a kid, but aren't quite sure what to name it? You create a Tinder-style webapp to help you figure it out!

..What? no? Well I did anyway.

This app allows you to import a set of names, and then mom & dad (or whichever configuration you prefer) can swipe on the names that pop up until they find some they both like. After going through the names the list can be reviewed and you can see which ones only one parent liked, and which ones both parents liked.

## Running it

It needs the following env vars set:

- `SITE_URL` points to the URL the server is running at (eg `http://localhost:5000/`).
- `COOKIE_SECRET` is the secret used encrypt cookies stored in the browser.
- `DATABASE_URL` is a valid PostgreSQL connection string/URL.
- `AUTH0_DOMAIN` is the Auth0 domain to be used for authentication.
- `OAUTH_CLIENT_ID` is the Auth0 OAuth client ID used for authentication.
- `OAUTH_CLIENT_CLIENT_SECRET` is the Auth0 OAuth client secret used for authentication.
- `DAD_EMAIL` is the e-mail address of the "dad" role.
- `MOM_EMAIL` is the e-mail address of the "mom" role.

Only emails matching `DAD_EMAIL`/`MOM_EMAIL` wil be let in.

## FAQ

1. What's the point?
   - Simple problems require over engineered solutions.

2. Should I use this?
   - Very likely not.

3. Why are there no tests?
   - Why are you still reading this? :(
