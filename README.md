# AnaMod for Tony Hawk's Project 8
#### v1.0.0

AnaMod is a comprehensive mod for Tony Hawk's Project 8 (both XBox 360 AND PS3) that works to enhance the base game with new features for speedrunners, but also adds a bunch of stuff for beginners and veterans of the game. 

## FEATURES
-    **Vanilla+**: Vanilla's physics and gameplay are untouched (outside of crash fixes and softlocks) - all of the new changes are behind menus, so you enable what you want!
-    **QoL Improvements**: Improvements are made to allow skipping every cutscene, allow the user to switch the back button (from Triangle/Y to Circle/B), disable Tips, disable filmer cameras, and much more!
-    **Achievements**: (PS3) All non-online achievements have been added and trigger the same way as they did on the 360 version. New speedrun achievements have been added!
-    **Speedrunning Practice**: Reworked Speedrun Overlay from THP8 Revamped with the ability to dynamically add new information like coordinates! AND, with the added ability to see what is needed to unlock all of the goals in the game!
-    **New Speedrun Mode**: Play through Any%, Rank 1, and 100% that follows the vanilla game's rules, but with the ability to use custom CAS', skip the first Tony cutscene, and see your load-removed time upon hitting Rank 8, 4, and 1!
-    **Time of Day**: Time of Day returns! Play through Project 8 on 24-minute and 48-minute days!
     -    Note: Lighting is hard, so it is mostly aesthetic :)
-    **Online Play Returns**: Long dormant, Online play is fixed and available for the PS3 version of the game.
     -    Note: This requires an RPCS3 patch. It is NOT required to use AnaMod, but will be if you wanna go online to play with friends!
     -    Second Note: For PS3, this is 100% something that will take time to nail down since a few some things didn't get ported over to the PS3 version.
     -    Third Note: There is jank.
-    **Crash Fixes**: Fixed some crashes and soft-locks found from speedrunning the game.
-    **More CAS Options**: Change your griptape, deck graphic, and wheel color from the CAS menu! Want to play THP8 as Tony Hawk? All pros are available for you to switch to at any time.

## REQUIREMENTS
-    Copy of THP8 PS3 or 360.
     -    PS3 and 360 game files must be decrypted!
     -    For 360, use [xdvdfs](https://github.com/antangelo/xdvdfs) to extract the 360 files.
-    RPCS3, Xenia (untested), RGH'd/Softmodded XBox 360, OR jailbroken PS3
-    Be a lover of THP8!!!!

## INSTALLATION
1.   Download AnaMod from the Releases tab based on your operating system.
2.   Run `AnaMod-Patcher-X` in its own folder
     -    Note: On Linux, you will need to use `chmod +x ./AnaMod-Patcher-Linux` to allow it to execute! Windows SmartScreen *may* prompt you about it, but this can be ignored.
     -    Once you run it, you will see a `patcher.log` and `anamod.ini` file.
          -    `patcher.log` gives more analysis of the patching process.
          -    `anamod.ini` keeps track of your THP8 folder location and your RPCS3 install location.
3.   Either drag your clean THP8 folder or click the box to open the file browser.
     -    Extracted PS3 folders have `PS3_GAME` and `PS3_DISC.SFB` at its root.
     -    Extracted 360 folders have `$SYSTEMUPDATE`, `DATA`, and `default.xex` at its root.
4.   If RPCS3, select the optional patches you want to install.
     -    RPCS3 must be closed, since the patcher will install the patches in the `Patch Manager` for THP8 and also checks them.
5.   Click `Update`.
     -    At any point, as long as `ANAMOD_BACKUP` is still in the folder structure, clicking `Uninstall` will revert the game back to its original state.
     -    Whenever an AnaMod update comes out, you do NOT need to Uninstall prior!

-    On 360: Copy the game folder over to your RGH'd/softmodded 360 via USB or FTP. Then, launch the game!
-    On PS3: Copy the game folder over to your PS3 and launch the game!
     -    NOTE: Current version of AnaMod does NOT apply AnaMod PS3 patches to the EBOOT.BIN directly. I'll work on this for the next release ^^
-    On RPCS3: Launch the game!

If the game didn't crash and you see `AnaMod v1.X` on the top-right of the main menu, you are good to go! :D

## PLANNED FEATURES
-    Revamped Integration
-    Enhanced Online Support (+server browser)
-    Individual Level Timer 
-    PS3 Console Support
     -    Already kinda know how to do this, but just not ready in time for v1!
-    More griptape options? Import deck graphics?????

## BUILDING
Building assumes you are on Linux and have Go `1.27.3` installed, and requires `anamod-data.zip` file from this repo.
-    `anamod-data.zip` is a package that holds the diffs of the AnaMod-edited scripts, images, and paks. A manifest file is included to point the patcher to what needs edited.

1.   `git clone https://github.com/TheAnastasis/AnaMod.git && cd AnaMod/patcher`
2.   Download `anamod-data.zip` from [Releases](https://github.com/TheAnastasis/AnaMod/releases/) and place it in `patcher/data/`.
3.   `go build -tags no_animations -o anamod-patcher .`
4.   `./anamod-patcher`

## CREDITS
-    **NodeQBC**
-    **Project 8 Revamped**