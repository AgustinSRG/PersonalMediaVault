
call heat dir "PersonalMediaVault" -gg -cg PComponentGroup -dr ProgramFiles64Folder -out HeatFile.wxs

call candle -ext WixUIExtension -ext WixUtilExtension -arch x64 HeatFile.wxs -o HeatFile.wixobj

call candle -ext WixUIExtension -ext WixUtilExtension -arch x64 Product.wxs -o Product.wixobj

call light Product.wixobj -spdb HeatFile.wixobj -b PersonalMediaVault -o PersonalMediaVault-1.2.0-x64.msi -ext WixUIExtension -ext WixUtilExtension
