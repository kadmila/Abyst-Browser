// core networking module for the Abyss browser (https://github.com/kadmila/Abyss-Browser)
//
// # Overview
// abyss-core development reference.
//
// **Build Profile** <br>
// There are two build profiles, debug or release. In build_dll_debug.ps1, -tags=debug enables debug profile.
//
// ## Conventions
//
// # Directory
// ## native_dll
// This directory includes DLL exports for abyssnet.dll <br>
// [dllmain.go](https://github.com/kadmila/Abyss-Browser/blob/main/abyss_core/native_dll/dllmain.go) contains the export definitions.
// marshalling.go provides dynamic length byte array marshalling.
//
// ## aerr
// Every error-derived classes must be defined here.
//
// ## ahmp
// Defines AHMP messages. For details, refer to [AHMP](AHMP). This does not provide serializer.
//
// ## and
// Abyss world neighbor discovery protocol implementation.
//
// ## aurl
//
// ## crash
//
// ## host
//
// ## interfaces
//
// ## net_service
//
// ## test
//
// ## test_logs
// .log files only; save test result here. This folder should not be pushed to the repository.
//
// ## tools
//
// ## watchdog
//
// Typical usage:
//
//	f := mypkg.NewFoo()
//	result := f.Process("example")
//
// For more details, see the individual type and function documentation.
package abysscore
