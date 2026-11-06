# v1.0.0 - The Initial Release

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
        -   Option only work on custom CAS' and not on Pro Skaters.
    -   Added the ability to choose your griptape.
        -   Option only work on custom CAS' and not on Pro Skaters.
    -   Added the ability to change your wheel colors.
        -   Option only work on custom CAS' and not on Pro Skaters.
  
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