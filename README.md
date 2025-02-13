# YouTube playlist counter

Counts the total duration of a youtube playlist with duration at different
speeds.

## usage

1. complie and build the binary

2. add a `.env` file with following key

    ```env
    YOUTUBE_API_KEY=
    ```

    the youtube api key should be **Youtube Data API v3**
    
3. run the program with first argument as playlist link.

    ```console
    ./ytplaycount <playlist_url>
    ```
