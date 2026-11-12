# v1.2.0
### Note: v1.2.0 or higher is/will required if you want to speedrun this game on SRC/thps.run.
-   Added
    -   Added a better check to stat and special trick resets.
    -   Added the ability to select Dawn, Noon, Dusk, and Midnight as your permanent time of day.

-   Fixed
    -   Fixed an issue where, if you had the achievement pop-ups on, the game would crash when you attain Rank 1 ([#3](https://github.com/TheAnastasis/AnaMod/issues/3)).
    -   Fixed an issue where a skater with a longer name than usual (Lyn-Z) would crash the player when you play online ([#5](https://github.com/TheAnastasis/AnaMod/issues/5)).
        -   Fixed an issue where the `PRESS X/A TO SKIP` UI element on the bottom could fail to be destroyed after a cutscene.
            -   This was a race condition on 360. That UI element is supposed to watch for X/A and skip the cutscene it spawned from. For at least the Unlock Factory goal, it technically finished the goal mid-cutscene due to how the triggers work. This essentially placed the script into a weird limbo state and kept eating X/A inputs forever. Cool.
    -   Fixed an issue where your stats would not be properly reset when you keep your skater between story modes.
    -   Fixed an issue where certain debug/career features (e.g. goal info) were not disabled in SPEEDRUN mode.
    -   Fixed an issue where Animal Chin's progress was no being cleared.
        -   Funnily enough, this is a vanilla bug and resulted in NG+. So, this kinda kills NG+ speedruns. So, if you still wanna do that, go with v1.1.2 lol
    -   Fixed an issue where the UI elements for competitions would render in th wrong location.

-   Changed
    -   Changed the time of day to start at noon/disabled when CAREER mode resets.
        -   SPEEDRUN mode is always set to Noon.
    -   Changed the Zombie Cut griptape so it no longer appears on the Xbox 360 version ([#6](https://github.com/TheAnastasis/AnaMod/issues/6)).
        -   Will work on this over time; it will continue to work on PS3 but is coded to not appear on 360.
    -   Changed the griptapes menu so that it will appear for pro skaters and CAS', even after initial creation.
***
# v1.1.2
-   Added
    -   Added the ability for the game to remember your time of day upon saving.
    -   Added new code that reduces the amount of pedestrians depending on time of day; 10 is the default (and maximum) at noon, with it gradually going down around dusk/dawn and until midnight when it drops to 3.
    -   Added the ability for players to edit their tricks Online when waiting for a new game to start.
    -   Added new UI elements for Competitions and Owned Spot Challenges to showcase score thresholds in CAREER mode (disabled in SPEEDRUN mode).
        -   Comps: Qualifiers show the required points to get into AM, PRO, and SICK; in the Competition, then it will show what you need to complete it (reminder there is ~1 point sway from RNG).
        -   Owned Spot Challenges: Shows the points required to get Sick (and of course you also need all objects touched).

-   Fixed
    -   Fixed an issue where the shuffle bonus text was appearing below the clean/sloppy text.
    -   Fixed an issue where HORSE was available in Online Play and caused a crash when it should've appeared in only Split Screen.
    -   Fixed an issue where you could not re-select `Beaver Valley` (aka the entire game world) if you switched levels in Online. 
    -   Fixed an issue where you could not edit special tricks when the `Edit Tricks` menu was avaialble in CAREER mode.
    -   Fixed an issue where the Achievement Unlocked text would still appear when the game is paused.
    -   Fixed an issue where the Time of Day suddenly changes for a few frames after certain cutscenes.
    -   Fixed an issue where the Lyn-Z/Nyjah Pro Challenge's UI would clip with the Speedrun Display.
        -   When the Speedrun Display is enabled, that scoreboard is moved to the right.
    -   Fixed an issue where the Speedrun Display would appear in certain cutscenes and in-game movies when it shouldn't.
        -   Added an additional check that, if a cutscene is playing, it will hide it. For movies, I just forgot to tell it to destroy the UI option.
    -   Fixed an issue with Ryan Sheckler's Pro Challenge failing to fully build the required steps to perform each step of the goal.
        -   This one I never knew of until Jarl did it by accident. In short, if you cancel out of one of Ryan's goals at a specific point, only portions of the goal's steps will appear at the top-right.
        -   The fix here is to just give you all of the steps at once, instead of trying to be cute and do it one by one.

-   Changed
    -   Changed the online behavior so you can launch matches when you are alone.
    -   Changed the online match options so it can only create "Player" lobbies.
        -   If online is figured out in the futrue, maybe this could open up again? Right now, it is was useless.
    -   Changed "IGT" to "LRT" in the Speedrun Display menu.
    -   Changed the Level Select order to the usual unlock progression and renamed them to their in-game names instead of their coded names (e.g. `CRETEPARK` -> `SKATE PARK`)
    -   Changed the `FPS Display` option in-game to appear on the top-left.

-   Removed
    -   Removed several Debug options that either do not work or they cause a crash.
        -   Full list: `Screen Calibration`, `Graphic Test`, and `Show Skater Pos` (already fulfilled by the Speedrun Display).
        -   Most of them, from what I can see, just are missing certain files.
    -   Removed the `CURRENT PLAYERS` option in Online play for now.
        -   This pretty much only works on Xbox Live - I'd have to figure out how to populate this, which will require probably another RPCS3 patch.
    -   Removed the Online Achievements, since there is no way for them to be collected right now.
        -   I meant to disable them before v1; will re-introduce them (and maybe some more) whenever the new Online update comes out!

***
# v1.0.0

## Main Game
-   Added
    -   Added a new option to skip un-skippable movies (e.g. Intro, Jason Lee) with the Ollie button.
        -   Also added a prompt to skip a cutscene if it is skippable.
            -   Only cutscenes not skippable are Jason Lee's in Speedrun Mode. The intro movie is skippable since it is before the time starts.
    -   Added a new option to skip all Tips.
        -   "Tips" given by goals (e.g. "Land it for..." and so on are still there).
    -   Added a new system to create loading screen messages!
    -   Added "Fast Rodney" back into the PS3 version.
        -   The PS3 version has an additional line of code not present in the 360 version; when removed, the trick works.
    -   (PS3) Added the ability to switch between Triangle and Circle as back buttons as an option.
        -   There is an optional patch that updates the UI to reflect this.
    -   Added a small notification to tell users they are entering a "NG+" mode.
        -   When enabled, Stokens, Achievements, Skate Shop unlocks, and Secret Skaters are kept.
    -   Added Day/Night cycles!
        -   This is experimental, and light sources don't work super well, so it is some sun manipulation and effects.
        -   In the options menu, you can choose "OFF", "24-MINUTE CYCLES", and "48-MINUTE CYCLES."
    -   Added a new career menu so you can choose between the classic STORY mode and the new SPEEDRUN mode (more below).
        -   Unfortunately, I could not figure out a good way to have two story progression states in one save. Flags could be separated, but getting goal unlocks and story progression to do the same would probably require a lot of default.xen or EBOOT.bin modifications... Which I ain't doing.
        -   If you reset to SPEEDRUN mode, it will wipe stokens and Skate Shop unlocks; CAREER mode resets weil carry these over.
    -   Added some entropy to song ordering when `Random` is selected.
  
-   Fixed
    -   Fixed an issue where, if you get 5 tricks/AM ranking within ~4 seconds on `First Time on Video` and hit `End Run`, the game would be put into a permanent limbo state, requriing a save and quit to proceed.
    -   Fixed an issue where the game could crash during `Out of Suburbia!` if the player spawned outside of the RC car's trigger as the cutscene ends.
        -   This one took a bit to track down, since it wasn't super obvious at first glance. Come to find out, the game hates it when you out of the area of an item being removed!
    -   Fixed an issue where the in-game timer began when you loaded into the bedroom instead of gaining control of your skater.
        -   Mostly to make IGT accurate.
  
## New CAS Options
-   Added
    -   Added the ability to switch to pro skaters as your CAS.
    -   Added the ability to choose your first deck graphic.
    -   Added the ability to choose your griptape.
    -   Added the ability to change your wheel colors.
  
## Achievements System
-   Added
    -   Added a new Achievements Manager that tracks all original 360 achievements and some custom ones!
        -   Also added some speedrun achievements!
    -   Added the ability to choose if achievement notifications appear in-game.
        -   Achievements showing up is disabled by default; you still gain achievements if the pop-up is offline!
  
## New Speedrunning Features
-   Added
    -   Added a new "Speedrun Display" UI element.
        -   This gives you real-time data on completion percentage, gaps collected, stoken count, and more!
        -   Also includes speed (in game units) and your X, Y, and Z coordinates!
        -   Notes:
            -   When enabled, the Stokens gained/lost indicator is removed.
            -   When enabled, certain ui elements are lowered to prevent UI overlaps.
            -   You also have a sub-menu to customize what UI elements you want to have appear!
            -   No matter the selections, most elements related to Career Mode (Goal Points, Rank, Gaps, etc.) are only available in Career Mode.
    -   Added a new option to skip all save prompts (remember: save often!).
    -   Added the ability to quickly reset runs.
        -   Runners also have the ability to change their skater too before returning to the game.
    -   Added a new option to skip all phone/pager prompts (Phone Skip!!).
        -   When enabled, it will skip all notifciations that can interrupt gameplay! You can still view them in the Pause Menu.
    -   Added a new UI element that shows you how many buttslaps you hit.
    -   Added a new "Goal Info" display that, when enabled, allows players to see the goal requirements (to include goal points required and which goals need to be completed) AND see how many goal points it rewards the player!

-   Speedrun Mode
    -   Enables a timer that displays upon completing Any%, Rank 1, and 100% (including ms)!
        -   Note: Time does not include loading screens.
    -   All Speedrun Display options are disabled.
    -   Most improvements to gameplay (e.g., phone skip) are disabled.
    -   Save prompts are disabled (for parity with 360 and allow custom CAS' without prompts)
    -   Day/Night cycles are disabled and forced to noon during speedruns.
    -   Jason Lee's cutscenes cannot be skipped.
    -   All cheats are disabled.
  
## THP8 Online (Experimental)
-   Restored online functionality for the PS3 version!
    -   As a note, there is a lot of things disabled in the engine. Because of that, if you enable the "Online" features, then this will require you to apply an RPCS3 patch, otherwise it won't appear!
        -   If you do not have the RPCS3 patch enabled, it will show the normal 2-Player screen!
    -   This enables direct connect! I do not know enough right now to add it in (though I have ideas). Maybe a future patch? :)
        -   Seriously, this took a lot longer than it should've x_x
    -   Connecting via IP will require the port to be forwarded (5150/UDP) - further support with OpenSpy and/or a custom option will be researched later.
        -   Also, expect some oddities if you are connecting to an IP with a high-ish ping for now.

-   Fixed
    -   Fixed an issue where a player will crash (even when they are host) when another play joins with their name.
        -   Added an additional check that, if you join and have the same name as another skater in the game, your connection is refused.
    -   Fixed an issue where the PS3 will not forward animation branches to other players.
        -   As you would expect, a lot just don't get forwarded since the original version never had it. Some things were brought over from the 360 version, but a lot of this was figured out after a lot of testing.
    -   Fixed an issue where assets with different states (e.g., the Slums shortcut to City Park, the building at the top of Skate Park, and the sewers next to School) would not load the proper state and cause voids in the world.
    -   Fixed an issue where animations could cancel themselves out and show nothing when latency is too high.